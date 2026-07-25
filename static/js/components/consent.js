const STORAGE_KEY = 'fixyourmoney-consent';
const CONSENT_LIFETIME = 1000 * 60 * 60 * 24 * 60;

/** @typedef {{externalMedia: boolean, timestamp: number}} Consent */

/** @type {Consent|null} */
let currentConsent = readConsent();

/**
 * Returns a defensive copy of the current consent state.
 * @returns {Consent|null}
 */
export function getConsent() {
  return currentConsent ? {...currentConsent} : null;
}

/**
 * Reads and validates the stored consent without failing when storage is blocked.
 * @returns {Consent|null}
 */
function readConsent() {
  try {
    const rawValue = window.localStorage.getItem(STORAGE_KEY);
    if (!rawValue) return null;
    const value = JSON.parse(rawValue);
    const validShape = typeof value.externalMedia === 'boolean' && typeof value.timestamp === 'number';
    if (!validShape || Date.now() - value.timestamp > CONSENT_LIFETIME) {
      window.localStorage.removeItem(STORAGE_KEY);
      return null;
    }
    return {externalMedia: value.externalMedia, timestamp: value.timestamp};
  } catch {
    return null;
  }
}

/**
 * Persists consent when possible and always updates the in-memory state.
 * @param {Consent} consent
 * @returns {void}
 */
function saveConsent(consent) {
  currentConsent = consent;
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(consent));
  } catch {
    // In-memory consent keeps this page usable when storage is unavailable.
  }
  window.dispatchEvent(new CustomEvent('consentchange', {detail: getConsent()}));
}

/**
 * Opens a native dialog without throwing when it is already open.
 * @param {HTMLDialogElement} dialog
 * @returns {void}
 */
function openDialog(dialog) {
  if (!dialog.open) dialog.showModal();
}

/**
 * Copies a consent state into the category controls.
 * @param {HTMLDialogElement} dialog
 * @param {Consent|null} consent
 * @returns {void}
 */
function syncControls(dialog, consent) {
  const externalMedia = dialog.querySelector('[data-consent-category="externalMedia"]');
  if (externalMedia instanceof HTMLInputElement) externalMedia.checked = Boolean(consent?.externalMedia);
}

/**
 * Reads the category controls from the dialog.
 * @param {HTMLDialogElement} dialog
 * @returns {{externalMedia: boolean}}
 */
function readControls(dialog) {
  const externalMedia = dialog.querySelector('[data-consent-category="externalMedia"]');
  return {
    externalMedia: externalMedia instanceof HTMLInputElement && externalMedia.checked,
  };
}

/**
 * Initializes consent controls, the 60-day expiry flow, and settings launchers.
 * @returns {void}
 */
export function initConsent() {
  const dialog = document.querySelector('[data-consent-dialog]');
  if (!(dialog instanceof HTMLDialogElement)) return;

  const launchers = document.querySelectorAll('[data-consent-open]');
  launchers.forEach((launcher) => {
    if (launcher instanceof HTMLElement) launcher.hidden = false;
    launcher.addEventListener('click', () => {
      syncControls(dialog, currentConsent);
      if (launcher.closest('[data-chart-panel]')) {
        const externalMedia = dialog.querySelector('[data-consent-category="externalMedia"]');
        if (externalMedia instanceof HTMLInputElement) externalMedia.checked = true;
      }
      openDialog(dialog);
    });
  });

  dialog.querySelector('[data-consent-close]')?.addEventListener('click', () => dialog.close());
  dialog.querySelectorAll('[data-consent-action]').forEach((button) => {
    button.addEventListener('click', () => {
      const action = button.getAttribute('data-consent-action');
      const selection = action === 'all'
        ? {externalMedia: true}
        : action === 'reject'
          ? {externalMedia: false}
          : readControls(dialog);
      saveConsent({...selection, timestamp: Date.now()});
      dialog.close();
    });
  });

  syncControls(dialog, currentConsent);
  if (!currentConsent) openDialog(dialog);
}

/**
 * Initializes native impact dialogs while retaining details fallbacks without JavaScript.
 * @returns {void}
 */
export function initDialogs() {
  document.querySelectorAll('[data-dialog-fallback]').forEach((fallback) => {
    if (fallback instanceof HTMLElement) fallback.hidden = true;
  });

  document.querySelectorAll('[data-dialog-open]').forEach((opener) => {
    if (!(opener instanceof HTMLButtonElement)) return;
    opener.hidden = false;
    opener.addEventListener('click', () => {
      const targetID = opener.getAttribute('data-dialog-open');
      const dialog = targetID ? document.getElementById(targetID) : null;
      if (!(dialog instanceof HTMLDialogElement)) return;
      dialog.showModal();
      dialog.addEventListener('close', () => opener.focus(), {once: true});
    });
  });

  document.querySelectorAll('[data-impact-dialog], [data-language-dialog]').forEach((dialog) => {
    if (!(dialog instanceof HTMLDialogElement)) return;
    dialog.querySelector('[data-dialog-close]')?.addEventListener('click', () => dialog.close());
  });
}

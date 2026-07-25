/**
 * Activates one tab and its associated panel.
 * @param {HTMLElement} tabList
 * @param {HTMLElement} nextTab
 * @param {boolean} moveFocus
 * @returns {void}
 */
function activateTab(tabList, nextTab, moveFocus) {
  const tabs = [...tabList.querySelectorAll('[role="tab"]')];
  const root = tabList.closest('[data-tabs]');
  if (!root) return;

  tabs.forEach((tab) => {
    const selected = tab === nextTab;
    tab.setAttribute('aria-selected', String(selected));
    tab.setAttribute('tabindex', selected ? '0' : '-1');
    const panelID = tab.getAttribute('aria-controls');
    const panel = panelID ? document.getElementById(panelID) : null;
    if (panel instanceof HTMLElement) panel.hidden = !selected;
  });
  if (moveFocus) nextTab.focus();
  root.dispatchEvent(new CustomEvent('tabchange', {detail: {tab: nextTab}}));
}

/**
 * Handles Arrow, Home, and End keys according to the ARIA tabs pattern.
 * @param {KeyboardEvent} event
 * @param {HTMLElement} tabList
 * @returns {void}
 */
function handleTabKeydown(event, tabList) {
  const tabs = [...tabList.querySelectorAll('[role="tab"]')].filter((tab) => tab instanceof HTMLElement);
  const currentIndex = tabs.indexOf(/** @type {HTMLElement} */ (event.currentTarget));
  if (currentIndex < 0) return;

  const direction = window.getComputedStyle(tabList).direction;
  const forward = direction === 'rtl' ? -1 : 1;
  let nextIndex = currentIndex;
  if (event.key === 'ArrowRight') nextIndex = (currentIndex + forward + tabs.length) % tabs.length;
  else if (event.key === 'ArrowLeft') nextIndex = (currentIndex - forward + tabs.length) % tabs.length;
  else if (event.key === 'ArrowDown') nextIndex = (currentIndex + 1) % tabs.length;
  else if (event.key === 'ArrowUp') nextIndex = (currentIndex - 1 + tabs.length) % tabs.length;
  else if (event.key === 'Home') nextIndex = 0;
  else if (event.key === 'End') nextIndex = tabs.length - 1;
  else return;

  event.preventDefault();
  activateTab(tabList, tabs[nextIndex], true);
}

/**
 * Initializes progressively enhanced tab groups; all panels remain readable without JavaScript.
 * @returns {void}
 */
export function initTabs() {
  document.querySelectorAll('[role="tablist"]').forEach((tabList) => {
    if (!(tabList instanceof HTMLElement)) return;
    const tabs = [...tabList.querySelectorAll('[role="tab"]')].filter((tab) => tab instanceof HTMLElement);
    const selected = tabs.find((tab) => tab.getAttribute('aria-selected') === 'true') || tabs[0];
    if (!selected) return;
    activateTab(tabList, selected, false);
    tabs.forEach((tab) => {
      tab.addEventListener('click', (event) => {
        event.preventDefault();
        activateTab(tabList, tab, false);
      });
      tab.addEventListener('keydown', (event) => handleTabKeydown(event, tabList));
    });
  });
}

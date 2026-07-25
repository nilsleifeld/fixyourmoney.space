const SECTION_SELECTOR = 'section[data-track-section][id]';
const SECTION_VIEW_THRESHOLD = 0.5;
const VIEWED_SECTION_EVENT = 'Viewed Section';

let sectionObserver = null;
let pendingSectionCount = 0;
let sectionTrackingInitialized = false;
const trackedSections = new Set();

/**
 * Sends a custom event through the Plausible client initialized by the layout.
 * Plausible queues the event while its remote script is still loading.
 * @param {string} eventName
 * @param {{props: Record<string, string>}} options
 * @returns {void}
 */
function trackEvent(eventName, options) {
  if (typeof window.plausible !== 'function') return;
  window.plausible(eventName, options);
}

/**
 * Stops tracking a section and releases the observer after the last section.
 * @param {Element} section
 * @param {IntersectionObserver} observer
 * @returns {void}
 */
function stopObservingSection(section, observer) {
  observer.unobserve(section);
  pendingSectionCount -= 1;

  if (pendingSectionCount === 0) {
    observer.disconnect();
    sectionObserver = null;
  }
}

/**
 * Reports sections once they are at least 50 percent visible.
 * @param {IntersectionObserverEntry[]} entries
 * @param {IntersectionObserver} observer
 * @returns {void}
 */
function handleSectionIntersections(entries, observer) {
  for (const entry of entries) {
    if (
      !entry.isIntersecting ||
      entry.intersectionRatio < SECTION_VIEW_THRESHOLD ||
      !(entry.target instanceof HTMLElement)
    ) {
      continue;
    }

    const section = entry.target.id;
    if (trackedSections.has(section)) {
      stopObservingSection(entry.target, observer);
      continue;
    }

    // Mark the section first so even a queued observer callback cannot report it twice.
    trackedSections.add(section);
    stopObservingSection(entry.target, observer);
    trackEvent(VIEWED_SECTION_EVENT, { props: { section } });
  }
}

/**
 * Initializes landing-page section tracking once for the current page view.
 * Sections opt in with `data-track-section` and use their ID as event property.
 * @returns {void}
 */
export function initSectionTracking() {
  if (sectionTrackingInitialized) return;
  sectionTrackingInitialized = true;

  if (!('IntersectionObserver' in window)) return;

  const sections = document.querySelectorAll(SECTION_SELECTOR);
  if (sections.length === 0) return;

  sectionObserver = new IntersectionObserver(handleSectionIntersections, {
    threshold: SECTION_VIEW_THRESHOLD,
  });
  pendingSectionCount = sections.length;

  for (const section of sections) {
    sectionObserver.observe(section);
  }
}

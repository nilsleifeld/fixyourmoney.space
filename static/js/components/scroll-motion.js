const REVEAL_SELECTOR = [
  '.content-section__lead',
  '.section-heading__eyebrow',
  '.section-heading__title',
  '.chart-panel',
  '.content-section__text',
  '.content-section__subtitle',
  '.point-list__item',
  '.money-card',
  '.money-card__icon',
  '.comparison',
  '.comparison__versus',
  '.impact-card',
  '.impact-card__icon',
  '.closing-message__title',
  '.legal-page__notice',
  '.legal-page__title',
  '.legal-page__section-title',
].join(',');

const LEFT_REVEAL_SELECTOR = [
  '.section-heading__eyebrow',
  '.money-card--tilt-left',
  '.point-list__item--from-left',
  '.comparison--from-left',
  '.impact-card--from-left',
].join(',');

const RIGHT_REVEAL_SELECTOR = [
  '.money-card--tilt-right',
  '.point-list__item--from-right',
  '.comparison--from-right',
  '.impact-card--from-right',
].join(',');

const SCALE_REVEAL_SELECTOR = [
  '.chart-panel',
  '.money-card__icon',
  '.comparison__versus',
  '.impact-card__icon',
  '.impact-card--from-bottom',
  '.closing-message__title',
].join(',');

/**
 * Chooses a contextual fallback motion based on the component variant.
 * @param {HTMLElement} element
 * @returns {'scroll-reveal--from-left'|'scroll-reveal--from-right'|'scroll-reveal--scale'|''}
 */
function revealModifier(element) {
  if (element.matches(LEFT_REVEAL_SELECTOR)) return 'scroll-reveal--from-left';
  if (element.matches(RIGHT_REVEAL_SELECTOR)) return 'scroll-reveal--from-right';
  if (element.matches(SCALE_REVEAL_SELECTOR)) return 'scroll-reveal--scale';
  return '';
}

/**
 * Marks one observed element as visible and stops observing it.
 * @param {HTMLElement} element
 * @param {IntersectionObserver} observer
 * @returns {void}
 */
function revealElement(element, observer) {
  element.classList.remove('scroll-reveal--pending');
  element.classList.add('scroll-reveal--visible');
  observer.unobserve(element);
}

/**
 * Reveals elements as they enter the viewport.
 * @param {IntersectionObserverEntry[]} entries
 * @param {IntersectionObserver} observer
 * @returns {void}
 */
function handleIntersections(entries, observer) {
  for (const entry of entries) {
    if (entry.isIntersecting && entry.target instanceof HTMLElement) {
      revealElement(entry.target, observer);
    }
  }
}

/**
 * Adds the contextual fallback reveal state and observes one element.
 * @param {HTMLElement} element
 * @param {IntersectionObserver} observer
 * @returns {void}
 */
function observeElement(element, observer) {
  const modifier = revealModifier(element);
  element.classList.add('scroll-reveal', 'scroll-reveal--pending');
  if (modifier) element.classList.add(modifier);
  observer.observe(element);
}

/**
 * Initializes a motion-safe fallback when native CSS view timelines are unavailable.
 * Browsers with view-timeline support use contextual scroll-linked CSS animations.
 * @returns {void}
 */
export function initScrollMotion() {
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const nativeScrollMotion = CSS.supports('animation-timeline: view()');
  if (reducedMotion || nativeScrollMotion || !('IntersectionObserver' in window)) return;

  const observer = new IntersectionObserver(handleIntersections, {
    rootMargin: '0px 0px -8% 0px',
    threshold: 0.08,
  });

  for (const element of document.querySelectorAll(REVEAL_SELECTOR)) {
    if (element instanceof HTMLElement) observeElement(element, observer);
  }
}

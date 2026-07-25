/**
 * Clears a pending exit animation and hides the surface when it finishes.
 * @param {HTMLElement} element
 * @param {string} activeClass
 * @param {string} exitClass
 * @returns {void}
 */
function finishHeroExit(element, activeClass, exitClass) {
  element.addEventListener(
    'animationend',
    (event) => {
      if (event.target !== element || element.classList.contains(activeClass)) return;
      element.classList.remove(exitClass);
      element.setAttribute('aria-hidden', 'true');
    },
    {once: true},
  );
}

/**
 * Activates or exits a hero surface with an optional flip animation.
 * @param {HTMLElement|null} element
 * @param {{ activeClass: string, exitClass: string, active: boolean, animate: boolean }} options
 * @returns {void}
 */
function setHeroSurfaceActive(element, options) {
  if (!(element instanceof HTMLElement)) return;
  const {activeClass, exitClass, active, animate} = options;

  if (!active) {
    if (!animate || !element.classList.contains(activeClass)) {
      element.classList.remove(activeClass, exitClass);
      element.setAttribute('aria-hidden', 'true');
      return;
    }
    element.classList.remove(activeClass);
    element.classList.add(exitClass);
    finishHeroExit(element, activeClass, exitClass);
    return;
  }

  element.classList.remove(exitClass);
  element.setAttribute('aria-hidden', 'false');
  if (!animate || element.classList.contains(activeClass)) {
    element.classList.add(activeClass);
    return;
  }
  // Force a style flush so adding the class always starts the CSS enter animation.
  void element.offsetWidth;
  element.classList.add(activeClass);
}

/**
 * Shows one hero chart without unloading the other (keeps TradingView mounts sized).
 * @param {HTMLElement|null} chart
 * @param {boolean} active
 * @param {boolean} animate
 * @returns {void}
 */
function setHeroChartActive(chart, active, animate) {
  setHeroSurfaceActive(chart, {
    activeClass: 'hero__chart--active',
    exitClass: 'hero__chart--exit',
    active,
    animate,
  });
  if (!(chart instanceof HTMLElement)) return;
  chart.toggleAttribute('inert', !active);
}

/**
 * Animates the title height so the toggle eases into its new vertical position.
 * @param {HTMLElement} title
 * @param {() => void} updateContent
 * @param {AbortController} controller
 * @returns {void}
 */
function animateTitleHeight(title, updateContent, controller) {
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const fromHeight = title.offsetHeight;
  title.style.height = `${fromHeight}px`;
  updateContent();

  const active = title.querySelector('.hero__perspective--active');
  if (!(active instanceof HTMLElement)) {
    title.style.height = '';
    return;
  }

  // Freeze the flip while measuring so rotate/scale cannot shrink the target height.
  active.style.setProperty('animation', 'none');
  const toHeight = active.offsetHeight;
  active.style.removeProperty('animation');
  active.classList.remove('hero__perspective--active');
  void active.offsetWidth;
  active.classList.add('hero__perspective--active');

  if (reduceMotion || fromHeight === toHeight) {
    title.style.height = '';
    return;
  }

  void title.offsetHeight;
  title.style.height = `${toHeight}px`;
  title.addEventListener(
    'transitionend',
    (event) => {
      if (event.target !== title || event.propertyName !== 'height') return;
      title.style.height = '';
    },
    {once: true, signal: controller.signal},
  );
}

/**
 * Updates the visible hero perspective and its accessible toggle state.
 * @param {HTMLElement} hero
 * @param {boolean} showSecond
 * @param {{ animate?: boolean, heightController?: AbortController }} [options]
 * @returns {void}
 */
function renderPerspective(hero, showSecond, options = {}) {
  const animate = options.animate === true;
  const first = hero.querySelector('[data-hero-perspective="first"]');
  const second = hero.querySelector('[data-hero-perspective="second"]');
  const firstLabel = hero.querySelector('[data-hero-label-first]');
  const secondLabel = hero.querySelector('[data-hero-label-second]');
  const firstChart = hero.querySelector('[data-hero-chart="first"]');
  const secondChart = hero.querySelector('[data-hero-chart="second"]');
  const toggle = hero.querySelector('[data-hero-toggle]');
  const title = hero.querySelector('.hero__title');

  if (animate) {
    hero.dataset.heroFlip = showSecond ? 'forward' : 'backward';
  } else {
    delete hero.dataset.heroFlip;
  }

  /**
   * Swaps the visible headline and related controls for the selected perspective.
   * @returns {void}
   */
  const updateContent = () => {
    setHeroSurfaceActive(first, {
      activeClass: 'hero__perspective--active',
      exitClass: 'hero__perspective--exit',
      active: !showSecond,
      animate,
    });
    setHeroSurfaceActive(second, {
      activeClass: 'hero__perspective--active',
      exitClass: 'hero__perspective--exit',
      active: showSecond,
      animate,
    });
    if (firstLabel instanceof HTMLElement) firstLabel.hidden = showSecond;
    if (secondLabel instanceof HTMLElement) secondLabel.hidden = !showSecond;
    setHeroChartActive(firstChart, !showSecond, animate);
    setHeroChartActive(secondChart, showSecond, animate);
    if (toggle instanceof HTMLElement) toggle.setAttribute('aria-pressed', String(showSecond));
  };

  if (animate && title instanceof HTMLElement && options.heightController instanceof AbortController) {
    animateTitleHeight(title, updateContent, options.heightController);
    return;
  }

  if (title instanceof HTMLElement) title.style.height = '';
  updateContent();
}

/**
 * Initializes the user-controlled hero perspective toggle.
 * @returns {void}
 */
export function initHero() {
  document.querySelectorAll('[data-hero]').forEach((hero) => {
    if (!(hero instanceof HTMLElement)) return;
    const toggle = hero.querySelector('[data-hero-toggle]');
    if (!(toggle instanceof HTMLButtonElement)) return;
    let showSecond = false;
    /** @type {AbortController|null} */
    let heightController = null;
    toggle.hidden = false;
    renderPerspective(hero, showSecond, {animate: false});
    toggle.addEventListener('click', () => {
      showSecond = !showSecond;
      heightController?.abort();
      heightController = new AbortController();
      renderPerspective(hero, showSecond, {animate: true, heightController});
    });
  });
}

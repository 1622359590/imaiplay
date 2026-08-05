const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
const observedReveal = new WeakSet<Element>();
const observedProgress = new WeakSet<Element>();
const observedCounter = new WeakSet<Element>();

export function initScrollReveal() {
  const observer = new IntersectionObserver((entries, owner) => entries.forEach((entry) => {
    if (entry.isIntersecting) { entry.target.classList.add('visible'); owner.unobserve(entry.target); }
  }), { threshold: .1, rootMargin: '0px 0px -40px 0px' });
  document.querySelectorAll('.stagger-group').forEach((group) => Array.from(group.children).forEach((child, index) => {
    (child as HTMLElement).style.transitionDelay = motionQuery.matches ? '0ms' : `${index * 80}ms`;
  }));
  document.querySelectorAll('.reveal').forEach((element) => {
    if (!observedReveal.has(element)) { observedReveal.add(element); observer.observe(element); }
  });
}

export function initProgressBars() {
  const observer = new IntersectionObserver((entries, owner) => entries.forEach((entry) => {
    if (entry.isIntersecting) { entry.target.classList.add('animate'); owner.unobserve(entry.target); }
  }), { threshold: .3 });
  document.querySelectorAll('.progress-fill').forEach((element) => {
    if (!observedProgress.has(element)) { observedProgress.add(element); observer.observe(element); }
  });
}

export function initNavbarScroll() {
  document.querySelectorAll('.app-header').forEach((header) => {
    if (header.getAttribute('data-scroll-bound')) return;
    header.setAttribute('data-scroll-bound', 'true');
    window.addEventListener('scroll', () => header.classList.toggle('scrolled', window.scrollY > 50), { passive: true });
  });
}

export function initCardTilt() {
  if (motionQuery.matches) return;
  document.querySelectorAll('.tilt-card').forEach((card) => {
    if (card.getAttribute('data-tilt-bound')) return;
    card.setAttribute('data-tilt-bound', 'true');
    card.addEventListener('mousemove', (event) => {
      const mouse = event as MouseEvent; const element = card as HTMLElement; const rect = element.getBoundingClientRect();
      const x = (mouse.clientX - rect.left) / rect.width - .5; const y = (mouse.clientY - rect.top) / rect.height - .5;
      element.style.transform = `perspective(600px) rotateY(${x * 6}deg) rotateX(${-y * 6}deg) translateY(-4px)`;
    });
    card.addEventListener('mouseleave', () => { (card as HTMLElement).style.transform = ''; });
  });
}

export function initCounters() {
  const observer = new IntersectionObserver((entries, owner) => entries.forEach((entry) => {
    if (!entry.isIntersecting) return;
    const element = entry.target as HTMLElement; const target = Number(element.dataset.count || 0);
    if (motionQuery.matches) { element.textContent = target.toLocaleString(); owner.unobserve(element); return; }
    const start = performance.now(); const tick = (now: number) => {
      const progress = Math.min((now - start) / 1200, 1); const eased = 1 - Math.pow(1 - progress, 3);
      element.textContent = Math.round(eased * target).toLocaleString();
      if (progress < 1) requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick); owner.unobserve(element);
  }), { threshold: .5 });
  document.querySelectorAll('[data-count]').forEach((element) => {
    if (!observedCounter.has(element)) { observedCounter.add(element); observer.observe(element); }
  });
}

export function initLearnerAnimations() {
  const run = () => { initScrollReveal(); initProgressBars(); initNavbarScroll(); initCardTilt(); initCounters(); };
  run();
  const mutationObserver = new MutationObserver(run);
  mutationObserver.observe(document.body, { childList: true, subtree: true });
  return () => mutationObserver.disconnect();
}

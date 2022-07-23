document.addEventListener('DOMContentLoaded', () => {
  const nav = document.querySelector('.sitenav');
  const navBody = nav.querySelector('.box > .body');
  const navList = navBody.querySelector('ul');
  const navPadding = 32; // >= navBody's non-collapsed padding

  // Toggle the navbox when the logo or anything in its title are clicked.
  const toggleNav = () => {
    // Animating height is a mess: https://stackoverflow.com/questions/3508605
    // When collapsing, set max-height to the actual height first so the
    // animation begins immediately. When expanding, set it to list's height
    // (plus extra for padding) so the animation takes roughly the right time.
    if (!nav.classList.contains('collapsed-mobile')) {
      navBody.style.maxHeight = navBody.clientHeight + 'px';
      window.setTimeout(() => (navBody.style.maxHeight = ''));
    } else {
      navBody.style.maxHeight = navList.clientHeight + navPadding + 'px';
    }
    nav.classList.toggle('collapsed-mobile');
  };
  document.querySelector('header .logo').addEventListener('click', toggleNav);
  document
    .querySelector('.sitenav .box .title')
    .addEventListener('click', toggleNav);

  // At the end of a transition, tell the body to use its natural height in case
  // the window is later resized.
  navBody.addEventListener('transitionend', () => {
    navBody.style.maxHeight = '';
  });

  // |darkQuery| and applyTheme() are defined in dark.js.
  // Toggle the theme when the dark-mode icon is clicked.
  // The initial state is set in base-body.js: we can't do this in the top level
  // of this file since document.body isn't available, and we also don't want to
  // do it in DOMContentLoaded since we'll get a flash of the light theme then.
  document
    .querySelector('header .dark')
    .addEventListener('click', () => applyTheme(true));

  // We may also need to update the theme if prefers-color-scheme changes.
  darkQuery.addEventListener('change', () => applyTheme());
});

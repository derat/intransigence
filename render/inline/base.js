document.addEventListener('DOMContentLoaded', () => {
  // Toggle the navbox when the logo or anything in its title are clicked.
  const toggle = () => {
    document.querySelector('nav ul').classList.toggle('collapsed-mobile');
    document.querySelector('nav .toggle').classList.toggle('expand');
  };
  document.querySelector('header .logo').addEventListener('click', toggle);
  document.querySelector('nav .box .title').addEventListener('click', toggle);

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

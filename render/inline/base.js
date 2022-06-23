document.addEventListener('DOMContentLoaded', () => {
  // Toggle the navbox when the logo or anything in its title are clicked.
  const toggle = () => {
    document.querySelector('nav ul').classList.toggle('collapsed-mobile');
    document.querySelector('nav .toggle').classList.toggle('expand');
  };
  document.querySelector('header .logo').addEventListener('click', toggle);
  document.querySelector('nav .box .title').addEventListener('click', toggle);

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

const darkQuery = window.matchMedia('(prefers-color-scheme: dark)');

// Adds or remove the 'dark' class from document.body per localStorage and
// prefers-color-scheme. If |toggle| is truthy, toggles the current value and
// saves the updated value to localStorage.
function applyTheme(toggle) {
  const hasStorage = typeof Storage !== 'undefined';
  let dark = false;
  if (toggle) {
    dark = !document.body.classList.contains('dark');
    if (hasStorage) localStorage.setItem('theme', dark ? 'dark' : 'light');
  } else {
    const saved = hasStorage ? localStorage.getItem('theme') : null;
    dark = saved !== null ? saved === 'dark' : darkQuery.matches;
  }
  dark
    ? document.body.classList.add('dark')
    : document.body.classList.remove('dark');
}

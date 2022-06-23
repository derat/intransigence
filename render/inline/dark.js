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

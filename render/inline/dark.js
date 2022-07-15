const darkQuery = window.matchMedia('(prefers-color-scheme: dark)');

// Adds or remove the 'dark' class from document.body per localStorage and
// prefers-color-scheme. If |toggle| is truthy, toggles the current value and
// saves the updated value to localStorage.
function applyTheme(toggle) {
  // AMP iframes can't use allow-same-origin since they might be served from the
  // cache. Check document.domain to determine if we're sandboxed, which
  // prevents us from accessing localStorage: https://stackoverflow.com/a/34073811
  //
  // Just give up and use the light theme in this case, since we won't be able
  // to tell if the user toggles the theme, and using the dark theme in an
  // iframe while the rest of the page is using the light theme looks weird.
  if (!document.domain) return;

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

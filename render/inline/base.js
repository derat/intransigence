document.addEventListener('DOMContentLoaded', () => {
  const toggle = () => {
    document.querySelector('nav ul').classList.toggle('collapsed-mobile');
    document.querySelector('nav .toggle').classList.toggle('expand');
  };
  document.querySelector('nav .logo').addEventListener('click', toggle);
  document.querySelector('nav .box .title').addEventListener('click', toggle);
});

// Wire up links to post messages to the iframe to activate markers.
document.addEventListener('DOMContentLoaded', () => {
  const iframe = document.getElementById('map');
  const anchors = document.getElementsByClassName('map-link');
  for (let i = 0; i < anchors.length; i++) {
    const a = anchors[i];
    const id = a.parentElement.parentElement.id;
    a.addEventListener('click', () =>
      iframe.contentWindow.postMessage({ id }, '*', [])
    );
  }
});

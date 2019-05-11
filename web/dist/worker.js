// This is required for 'add to homescreen' to work.
this.addEventListener('fetch', function(event) {
  console.log('fetch', event);
  // event.respondWith(fetch(event.request));
});

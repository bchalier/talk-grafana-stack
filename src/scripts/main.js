var isWebKit = 'webkitAppearance' in document.documentElement.style,
  // zoom-based scaling causes font sizes and line heights to be calculated differently
  // on the other hand, zoom-based scaling correctly anti-aliases fonts during transforms (no need for layer creation hack)
  scaleMethod = isWebKit ? 'zoom' : 'transform',
  bespoke = require('bespoke'),
  classes = require('bespoke-classes'),
  nav = require('bespoke-nav'),
  fullscreen = require('bespoke-fullscreen'),
  scale = require('bespoke-scale'),
  overview = require('bespoke-overview'),
  bullets = require('bespoke-bullets'),
  hash = require('bespoke-hash'),
  prism = require('bespoke-prism'),
  multimedia = require('bespoke-multimedia'),
  extern = require('bespoke-extern');

bespoke.from({ parent: 'article.deck', slides: 'section' }, [
  classes(),
  nav(),
  fullscreen(),
  scale(scaleMethod),
  overview({ columns: 4 }),
  bullets('.build, .build-items > *:not(.build-items)'),
  hash(),
  prism(),
  multimedia(),
  extern(bespoke)
]);

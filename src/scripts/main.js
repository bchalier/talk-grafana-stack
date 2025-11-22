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
  hash = require('bespoke-hash'),
  prism = require('bespoke-prism'),
  multimedia = require('bespoke-multimedia'),
  extern = require('bespoke-extern');

const manualBullets = (options) => (deck) => {
  const selector = typeof options === 'string' ? options : '[data-bespoke-bullet]';
  const bullets = deck.slides.map((slide) =>
    [].slice.call(slide.querySelectorAll(selector), 0)
  );

  let activeSlideIndex = 0;
  let activeBulletIndex = -1;

  const updateFocus = () => {
    deck.slides.forEach((slide) => slide.classList.remove('focus-kube'));
    const current = bullets[activeSlideIndex][activeBulletIndex];
    if (current && current.dataset && current.dataset.focus === 'kube') {
      deck.slides[activeSlideIndex].classList.add('focus-kube');
    }
  };

  const activateBullet = (slideIndex, bulletIndex) => {
    activeSlideIndex = slideIndex;
    activeBulletIndex = bulletIndex;

    bullets.forEach((slide, s) => {
      slide.forEach((bullet, b) => {
        bullet.classList.add('bespoke-bullet');

        const shouldBeActive = s < slideIndex || (s === slideIndex && b <= bulletIndex && bulletIndex >= 0);

        if (shouldBeActive) {
          bullet.classList.add('bespoke-bullet-active');
          bullet.classList.remove('bespoke-bullet-inactive');
        } else {
          bullet.classList.add('bespoke-bullet-inactive');
          bullet.classList.remove('bespoke-bullet-active');
        }

        if (s === slideIndex && b === bulletIndex) {
          bullet.classList.add('bespoke-bullet-current');
        } else {
          bullet.classList.remove('bespoke-bullet-current');
        }
      });
    });

    updateFocus();
  };

  const activeSlideHasBulletByOffset = (offset) =>
    bullets[activeSlideIndex][activeBulletIndex + offset] !== undefined;

  const next = () => {
    const nextSlideIndex = activeSlideIndex + 1;

    if (activeSlideHasBulletByOffset(1)) {
      activateBullet(activeSlideIndex, activeBulletIndex + 1);
      return false;
    } else if (bullets[nextSlideIndex]) {
      activateBullet(nextSlideIndex, -1);
    }
  };

  const prev = () => {
    const prevSlideIndex = activeSlideIndex - 1;

    if (activeBulletIndex >= 0 && activeSlideHasBulletByOffset(-1)) {
      activateBullet(activeSlideIndex, activeBulletIndex - 1);
      return false;
    } else if (bullets[prevSlideIndex]) {
      activateBullet(prevSlideIndex, bullets[prevSlideIndex].length - 1);
    }
  };

  deck.on('next', next);
  deck.on('prev', prev);

  deck.on('slide', (e) => {
    activateBullet(e.index, -1);
  });

  activateBullet(0, -1);
};

bespoke.from({ parent: 'article.deck', slides: 'section' }, [
  classes(),
  nav(),
  fullscreen(),
  scale(scaleMethod),
  overview({ columns: 4 }),
  manualBullets('.build, .build-items > *:not(.build-items)'),
  hash(),
  prism(),
  multimedia(),
  extern(bespoke)
]);

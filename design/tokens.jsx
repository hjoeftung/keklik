// Keklik design tokens — two palette directions ("pastel" and "bright")
// Both warm, both family-friendly. Pastel is the default.

const PALETTES = {
  pastel: {
    name: 'Pastel',
    bg: '#FBF6EE',
    surface: '#FFFFFF',
    surfaceSoft: '#F4ECDD',
    ink: '#2E2A33',
    inkSoft: '#6E6776',
    inkMuted: '#9A93A2',
    border: '#EADFCB',
    primary: '#E59B6A',      // warm apricot — primary CTA
    primaryDeep: '#C97B4D',
    primarySoft: '#FBE4D1',
    night: '#5B7BB8',        // night sleep accent — soft denim
    nightSoft: '#D9E2F3',
    nightDeep: '#3F5A8E',
    nap: '#E8B86E',          // nap accent — honey
    napSoft: '#F8E4B8',
    accent: '#A38FC4',       // gentle lavender for highlights
    accentSoft: '#E5DCF2',
    success: '#86B6A6',
    danger: '#D4806E',
    moon: '#FFE9B0',
    nightSky: '#2C2745',
    nightSkyDeep: '#1B1733',
  },
  bright: {
    name: 'Bright',
    bg: '#FFF4E6',
    surface: '#FFFFFF',
    surfaceSoft: '#FFE8C9',
    ink: '#2A1F12',
    inkSoft: '#6B5638',
    inkMuted: '#A48A66',
    border: '#F2D9B6',
    primary: '#FF8C5A',      // saturated coral
    primaryDeep: '#E2683A',
    primarySoft: '#FFD4B8',
    night: '#5566D8',        // bright periwinkle
    nightSoft: '#D5DAFF',
    nightDeep: '#3A47B0',
    nap: '#F7B924',          // marigold
    napSoft: '#FFE5A6',
    accent: '#C26ACF',       // playful magenta
    accentSoft: '#F2D7F8',
    success: '#5FBF8E',
    danger: '#E55B4F',
    moon: '#FFD86E',
    nightSky: '#241F4D',
    nightSkyDeep: '#15123A',
  },
};

const TYPE = {
  // Quicksand: rounded geometric, warm and friendly
  // Fraunces: a touch of personality for display
  // Use Quicksand only — keep type system tight
  family: '"Quicksand", "Nunito", "SF Pro Rounded", -apple-system, system-ui, sans-serif',
  display: '"Fraunces", "Quicksand", Georgia, serif',
};

// Inject base styles + Google Fonts once
if (typeof document !== 'undefined' && !document.getElementById('keklik-base')) {
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = 'https://fonts.googleapis.com/css2?family=Quicksand:wght@400;500;600;700&family=Fraunces:opsz,wght@9..144,400;9..144,500;9..144,600&display=swap';
  document.head.appendChild(link);

  const s = document.createElement('style');
  s.id = 'keklik-base';
  s.textContent = `
    .kk { font-family: ${TYPE.family}; -webkit-font-smoothing: antialiased; color: var(--ink); }
    .kk * { box-sizing: border-box; }
    .kk-display { font-family: ${TYPE.display}; font-weight: 500; letter-spacing: -0.01em; }
    .kk-num { font-variant-numeric: tabular-nums; font-feature-settings: "tnum"; }
    @keyframes kk-breathe {
      0%, 100% { transform: scale(1); }
      50% { transform: scale(1.035); }
    }
    @keyframes kk-float {
      0%, 100% { transform: translateY(0); }
      50% { transform: translateY(-4px); }
    }
    @keyframes kk-twinkle {
      0%, 100% { opacity: 0.3; }
      50% { opacity: 1; }
    }
    @keyframes kk-zfloat {
      0% { transform: translate(0, 0) scale(0.8); opacity: 0; }
      30% { opacity: 1; }
      100% { transform: translate(20px, -40px) scale(1.2); opacity: 0; }
    }
    .kk-scroll { overflow-y: auto; scrollbar-width: none; }
    .kk-scroll::-webkit-scrollbar { display: none; }
  `;
  document.head.appendChild(s);
}

Object.assign(window, { PALETTES, TYPE });

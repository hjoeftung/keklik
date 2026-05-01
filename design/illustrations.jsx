// Hand-drawn illustrations — pillows, moons, clouds, stars, Z's.
// Each accepts a `palette` and renders inline SVG. Style: soft fills,
// gentle off-axis highlights, no harsh outlines.

function Pillow({ size = 200, palette, sleeping = false, tilt = 0 }) {
  const p = palette;
  const fill = p.primarySoft;
  const shade = p.primary;
  return (
    <svg width={size} height={size} viewBox="0 0 200 160" style={{ transform: `rotate(${tilt}deg)`, transition: 'transform .4s cubic-bezier(.4,0,.2,1)' }}>
      <defs>
        <radialGradient id={`pil-${p.name}`} cx="35%" cy="30%">
          <stop offset="0%" stopColor="#fff" stopOpacity="0.9" />
          <stop offset="60%" stopColor={fill} stopOpacity="0" />
        </radialGradient>
      </defs>
      {/* main pillow body — soft squircle with two corner dimples */}
      <path
        d="M 30 30
           Q 30 15, 50 18
           Q 100 8, 150 18
           Q 170 15, 170 30
           Q 178 80, 170 130
           Q 170 145, 150 142
           Q 100 152, 50 142
           Q 30 145, 30 130
           Q 22 80, 30 30 Z"
        fill={fill}
        stroke={shade}
        strokeWidth="2.5"
        strokeOpacity="0.18"
      />
      {/* highlight */}
      <path
        d="M 30 30 Q 30 15, 50 18 Q 100 8, 150 18 Q 170 15, 170 30"
        fill="none"
        stroke="#fff"
        strokeWidth="3"
        strokeLinecap="round"
        strokeOpacity="0.55"
      />
      {/* corner dimples */}
      <circle cx="40" cy="32" r="2.5" fill={shade} opacity="0.25" />
      <circle cx="160" cy="32" r="2.5" fill={shade} opacity="0.25" />
      <circle cx="40" cy="128" r="2.5" fill={shade} opacity="0.25" />
      <circle cx="160" cy="128" r="2.5" fill={shade} opacity="0.25" />
      {/* gentle gradient sheen */}
      <ellipse cx="80" cy="55" rx="45" ry="22" fill={`url(#pil-${p.name})`} />
      {sleeping && (
        <g>
          <text x="120" y="60" fontSize="22" fontFamily="Quicksand, sans-serif" fontWeight="700" fill={p.night} opacity="0.6">z</text>
          <text x="138" y="48" fontSize="16" fontFamily="Quicksand, sans-serif" fontWeight="700" fill={p.night} opacity="0.5">z</text>
        </g>
      )}
    </svg>
  );
}

function PillowFlat({ size = 200, palette, label, dimples = true }) {
  const p = palette;
  return (
    <svg width={size} height={size * 0.8} viewBox="0 0 200 160">
      <path
        d="M 28 32
           Q 28 16, 48 18
           Q 100 9, 152 18
           Q 172 16, 172 32
           Q 180 80, 172 128
           Q 172 144, 152 142
           Q 100 151, 48 142
           Q 28 144, 28 128
           Q 20 80, 28 32 Z"
        fill={p.primary}
        stroke={p.primaryDeep}
        strokeWidth="2"
        strokeOpacity="0.2"
      />
      <path
        d="M 32 32 Q 32 22, 50 24 Q 100 16, 150 24 Q 168 22, 168 32"
        fill="none" stroke="#fff" strokeWidth="3" strokeLinecap="round" strokeOpacity="0.4"
      />
      {dimples && <>
        <circle cx="42" cy="34" r="2.2" fill={p.primaryDeep} opacity="0.35" />
        <circle cx="158" cy="34" r="2.2" fill={p.primaryDeep} opacity="0.35" />
        <circle cx="42" cy="126" r="2.2" fill={p.primaryDeep} opacity="0.35" />
        <circle cx="158" cy="126" r="2.2" fill={p.primaryDeep} opacity="0.35" />
      </>}
      {label && (
        <text x="100" y="92" textAnchor="middle" fontSize="22" fontFamily="Quicksand, sans-serif" fontWeight="600" fill="#fff">
          {label}
        </text>
      )}
    </svg>
  );
}

// Pillow with a sleep mask draped over it — for one of the active states
function PillowMasked({ size = 200, palette }) {
  const p = palette;
  return (
    <svg width={size} height={size * 0.8} viewBox="0 0 200 160">
      <path
        d="M 28 32 Q 28 16, 48 18 Q 100 9, 152 18 Q 172 16, 172 32
           Q 180 80, 172 128 Q 172 144, 152 142 Q 100 151, 48 142
           Q 28 144, 28 128 Q 20 80, 28 32 Z"
        fill={p.primarySoft}
        stroke={p.primaryDeep}
        strokeWidth="2"
        strokeOpacity="0.2"
      />
      {/* sleep mask — tilted band across the pillow */}
      <g transform="rotate(-8 100 80)">
        <path
          d="M 20 65 Q 100 55, 180 65 Q 180 95, 100 92 Q 20 95, 20 65 Z"
          fill={p.night}
        />
        {/* eye stitches — closed eyes */}
        <path d="M 60 78 Q 70 84, 80 78" fill="none" stroke="#fff" strokeWidth="2.4" strokeLinecap="round" />
        <path d="M 120 78 Q 130 84, 140 78" fill="none" stroke="#fff" strokeWidth="2.4" strokeLinecap="round" />
        {/* strap */}
        <path d="M 20 75 Q 4 80, 4 88" fill="none" stroke={p.nightDeep} strokeWidth="2.5" strokeLinecap="round" opacity="0.4" />
        <path d="M 180 75 Q 196 80, 196 88" fill="none" stroke={p.nightDeep} strokeWidth="2.5" strokeLinecap="round" opacity="0.4" />
      </g>
    </svg>
  );
}

function Moon({ size = 60, palette, mode = 'crescent' }) {
  const p = palette;
  if (mode === 'full') {
    return (
      <svg width={size} height={size} viewBox="0 0 60 60">
        <circle cx="30" cy="30" r="22" fill={p.moon} />
        <circle cx="22" cy="24" r="4" fill={p.primaryDeep} opacity="0.12" />
        <circle cx="38" cy="34" r="2.5" fill={p.primaryDeep} opacity="0.15" />
        <circle cx="32" cy="42" r="2" fill={p.primaryDeep} opacity="0.1" />
      </svg>
    );
  }
  return (
    <svg width={size} height={size} viewBox="0 0 60 60">
      <path d="M 42 12 A 22 22 0 1 0 42 48 A 16 16 0 1 1 42 12 Z" fill={p.moon} />
    </svg>
  );
}

function Cloud({ size = 100, palette, dark = false }) {
  const p = palette;
  const fill = dark ? p.night : p.surface;
  const stroke = dark ? p.nightDeep : p.border;
  return (
    <svg width={size} height={size * 0.6} viewBox="0 0 100 60">
      <path
        d="M 22 42 Q 8 42, 8 32 Q 8 22, 20 22 Q 22 12, 36 12 Q 48 8, 56 18 Q 70 16, 74 28 Q 88 28, 88 38 Q 88 50, 76 50 L 26 50 Q 18 50, 22 42 Z"
        fill={fill}
        stroke={stroke}
        strokeWidth="2"
        strokeOpacity={dark ? 0.3 : 1}
      />
    </svg>
  );
}

function Star({ size = 16, palette, twinkle = false, delay = 0 }) {
  const p = palette;
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" style={twinkle ? { animation: `kk-twinkle 2.4s ease-in-out infinite ${delay}s` } : {}}>
      <path
        d="M 8 1 L 9.4 6.6 L 15 8 L 9.4 9.4 L 8 15 L 6.6 9.4 L 1 8 L 6.6 6.6 Z"
        fill={p.moon}
      />
    </svg>
  );
}

// A small night-sky scene used behind active sleep state
function NightScene({ width = 358, height = 200, palette }) {
  const p = palette;
  return (
    <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`} style={{ display: 'block' }}>
      <defs>
        <linearGradient id={`sky-${p.name}`} x1="0" x2="0" y1="0" y2="1">
          <stop offset="0%" stopColor={p.nightSky} />
          <stop offset="100%" stopColor={p.nightSkyDeep} />
        </linearGradient>
      </defs>
      <rect width={width} height={height} fill={`url(#sky-${p.name})`} />
      {/* stars */}
      {Array.from({ length: 18 }).map((_, i) => {
        const x = (i * 53) % width + 12;
        const y = (i * 37) % (height - 60) + 16;
        const r = (i % 3) * 0.6 + 0.8;
        return <circle key={i} cx={x} cy={y} r={r} fill={p.moon} opacity={0.4 + (i % 4) * 0.15} />;
      })}
      {/* moon */}
      <g transform={`translate(${width - 80}, 30)`}>
        <circle cx="20" cy="20" r="22" fill={p.moon} opacity="0.15" />
        <circle cx="20" cy="20" r="16" fill={p.moon} />
        <circle cx="14" cy="16" r="2.6" fill={p.primaryDeep} opacity="0.15" />
        <circle cx="24" cy="22" r="1.8" fill={p.primaryDeep} opacity="0.18" />
      </g>
      {/* far cloud */}
      <path
        d={`M 20 ${height - 30} Q 60 ${height - 50}, 110 ${height - 32} Q 160 ${height - 48}, 220 ${height - 30} Q 280 ${height - 22}, ${width} ${height - 28} L ${width} ${height} L 0 ${height} Z`}
        fill={p.nightSkyDeep}
        opacity="0.6"
      />
    </svg>
  );
}

// Tiny baby/bunny mascot — used sparsely (settings, empty states)
function Bun({ size = 80, palette, sleeping = false }) {
  const p = palette;
  return (
    <svg width={size} height={size} viewBox="0 0 80 80">
      {/* ears */}
      <ellipse cx="28" cy="20" rx="6" ry="14" fill={p.primarySoft} stroke={p.primary} strokeWidth="1.5" strokeOpacity="0.4" />
      <ellipse cx="52" cy="20" rx="6" ry="14" fill={p.primarySoft} stroke={p.primary} strokeWidth="1.5" strokeOpacity="0.4" />
      <ellipse cx="28" cy="22" rx="2.4" ry="7" fill={p.primary} opacity="0.4" />
      <ellipse cx="52" cy="22" rx="2.4" ry="7" fill={p.primary} opacity="0.4" />
      {/* head */}
      <circle cx="40" cy="46" r="22" fill={p.primarySoft} stroke={p.primary} strokeWidth="1.5" strokeOpacity="0.4" />
      {/* cheeks */}
      <circle cx="28" cy="52" r="3.5" fill={p.primary} opacity="0.35" />
      <circle cx="52" cy="52" r="3.5" fill={p.primary} opacity="0.35" />
      {/* eyes */}
      {sleeping ? (
        <>
          <path d="M 30 44 Q 34 47, 38 44" fill="none" stroke={p.ink} strokeWidth="1.8" strokeLinecap="round" />
          <path d="M 42 44 Q 46 47, 50 44" fill="none" stroke={p.ink} strokeWidth="1.8" strokeLinecap="round" />
        </>
      ) : (
        <>
          <circle cx="34" cy="44" r="2" fill={p.ink} />
          <circle cx="46" cy="44" r="2" fill={p.ink} />
        </>
      )}
      {/* nose */}
      <path d="M 40 50 L 38 52 L 42 52 Z" fill={p.primaryDeep} opacity="0.6" />
      {/* mouth */}
      <path d="M 40 52 Q 38 56, 36 56 M 40 52 Q 42 56, 44 56" fill="none" stroke={p.ink} strokeWidth="1.4" strokeLinecap="round" />
    </svg>
  );
}

Object.assign(window, { Pillow, PillowFlat, PillowMasked, Moon, Cloud, Star, NightScene, Bun });

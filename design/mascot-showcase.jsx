// Kek mascot showcase — portrait, states, icons, and pillow interactions

// ─── Kek bird — main character portrait ────────────────────────────────
function KekBird({ palette, state = 'awake', size = 120 }) {
  const p = palette;
  const s = size;
  
  // Eye state
  const eyeStates = {
    awake: { pupilX: 0, pupilY: 0, lashOpacity: 1 },
    peek: { pupilX: -2, pupilY: -3, lashOpacity: 0.6 },
    sleep: { pupilX: 0, pupilY: 5, lashOpacity: 0 },
  };
  const eye = eyeStates[state] || eyeStates.awake;

  return (
    <svg width={s} height={s} viewBox="0 0 120 140" style={{ overflow: 'visible' }}>
      {/* Body — round fluffy keklik shape */}
      <ellipse cx="60" cy="75" rx="38" ry="42" fill={p.primary} />
      
      {/* Subtle texture — ruffled feathers on sides */}
      <ellipse cx="28" cy="65" rx="8" ry="20" fill={p.primaryDeep} opacity="0.15" />
      <ellipse cx="92" cy="65" rx="8" ry="20" fill={p.primaryDeep} opacity="0.15" />
      
      {/* Wing shape (soft) */}
      <path d="M 30 75 Q 20 80, 22 95 Q 28 90, 30 85 Z" fill={p.primaryDeep} opacity="0.2" />
      <path d="M 90 75 Q 100 80, 98 95 Q 92 90, 90 85 Z" fill={p.primaryDeep} opacity="0.2" />
      
      {/* Head */}
      <circle cx="60" cy="50" r="28" fill={p.primary} />
      
      {/* Tuft — the stubborn one */}
      <ellipse cx="60" cy="16" rx="6" ry="14" fill={p.primary} />
      <ellipse cx="55" cy="12" rx="5" ry="12" fill={p.primaryDeep} opacity="0.3" />
      
      {/* Eye sockets */}
      <circle cx="48" cy="42" r="8" fill={p.primaryDeep} opacity="0.2" />
      <circle cx="72" cy="42" r="8" fill={p.primaryDeep} opacity="0.2" />
      
      {/* Eyes — white with pupil */}
      <circle cx="48" cy="42" r="6" fill="#fff" />
      <circle cx="72" cy="42" r="6" fill="#fff" />
      
      {/* Pupils */}
      <circle cx={48 + eye.pupilX} cy={42 + eye.pupilY} r="3.5" fill={p.night} />
      <circle cx={72 + eye.pupilX} cy={42 + eye.pupilY} r="3.5" fill={p.night} />
      
      {/* Lashes (only when awake) */}
      {eye.lashOpacity > 0 && (
        <g opacity={eye.lashOpacity}>
          <line x1="48" y1="36" x2="48" y2="31" stroke={p.primaryDeep} strokeWidth="1.2" strokeLinecap="round" />
          <line x1="45" y1="35" x2="42" y2="31" stroke={p.primaryDeep} strokeWidth="1" strokeLinecap="round" />
          <line x1="51" y1="35" x2="54" y2="31" stroke={p.primaryDeep} strokeWidth="1" strokeLinecap="round" />
          <line x1="72" y1="36" x2="72" y2="31" stroke={p.primaryDeep} strokeWidth="1.2" strokeLinecap="round" />
          <line x1="69" y1="35" x2="66" y2="31" stroke={p.primaryDeep} strokeWidth="1" strokeLinecap="round" />
          <line x1="75" y1="35" x2="78" y2="31" stroke={p.primaryDeep} strokeWidth="1" strokeLinecap="round" />
        </g>
      )}
      
      {/* Beak — honey colored */}
      <path d="M 60 52 L 68 58 L 60 61 Z" fill={p.nap} />
      <path d="M 60 52 L 52 58 L 60 61 Z" fill={p.nap} opacity="0.7" />
      
      {/* Blush */}
      <circle cx="38" cy="60" r="6" fill={p.nap} opacity="0.3" />
      <circle cx="82" cy="60" r="6" fill={p.nap} opacity="0.3" />
      
      {/* Feet */}
      <g>
        <line x1="52" y1="115" x2="52" y2="125" stroke={p.nap} strokeWidth="2" strokeLinecap="round" />
        <line x1="68" y1="115" x2="68" y2="125" stroke={p.nap} strokeWidth="2" strokeLinecap="round" />
        <circle cx="52" cy="128" r="2.5" fill={p.nap} />
        <circle cx="68" cy="128" r="2.5" fill={p.nap} />
      </g>
    </svg>
  );
}

// ─── Showcase: Kek portrait ────────────────────────────────────────────
function KekShowcase({ palette }) {
  const p = palette;
  return (
    <div style={{
      background: p.bg,
      height: '100%',
      padding: 40,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: '"Quicksand", system-ui',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: 24 }}>
        <KekBird palette={p} state="awake" size={200} />
      </div>
      <div className="kk-display" style={{ fontSize: 24, fontWeight: 500, color: p.ink, textAlign: 'center', letterSpacing: -0.4 }}>
        Kek
      </div>
      <div style={{ fontSize: 14, color: p.inkSoft, textAlign: 'center', marginTop: 8, lineHeight: 1.5, maxWidth: 280 }}>
        Round, chubby, slightly ruffled. Warm cream body; honey beak; one stubborn tuft.
      </div>
    </div>
  );
}

// ─── Showcase: Kek states ─────────────────────────────────────────────
function KekStates({ palette }) {
  const p = palette;
  const states = [
    { label: 'Awake', state: 'awake' },
    { label: 'Peek', state: 'peek' },
    { label: 'Sleep', state: 'sleep' },
  ];
  
  return (
    <div style={{
      background: p.bg,
      height: '100%',
      padding: 32,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: '"Quicksand", system-ui',
    }}>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 24, alignItems: 'center', justifyContent: 'center' }}>
        {states.map(({ label, state }) => (
          <div key={state} style={{ textAlign: 'center' }}>
            <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 12 }}>
              <KekBird palette={p} state={state} size={100} />
            </div>
            <div style={{ fontSize: 13, fontWeight: 600, color: p.ink }}>{label}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Simple app icon directions ─────────────────────────────────────────
function IconShowcase({ palette }) {
  const p = palette;
  return (
    <div style={{
      background: p.bg,
      height: '100%',
      padding: 32,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: '"Quicksand", system-ui',
    }}>
      <div style={{ display: 'flex', gap: 32, alignItems: 'center' }}>
        {/* Icon option 1: Kek portrait on pillow background */}
        <div style={{
          width: 120,
          height: 120,
          borderRadius: 28,
          background: `linear-gradient(135deg, ${p.primary}, ${p.primarySoft})`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: `0 12px 24px ${p.primary}30`,
        }}>
          <KekBird palette={p} state="awake" size={80} />
        </div>
        
        {/* Icon option 2: Just pillow with subtle kek hint */}
        <div style={{
          width: 120,
          height: 120,
          borderRadius: 28,
          background: p.primary,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: `0 12px 24px ${p.primary}30`,
        }}>
          <PillowFlat palette={p} size={100} />
        </div>
      </div>
      <div style={{ fontSize: 13, color: p.inkSoft, marginTop: 20, textAlign: 'center' }}>
        App icon directions · Kek on pillow or pillow alone
      </div>
    </div>
  );
}

// ─── Kek on pillow showcase (the active state interaction) ──────────────
function KekOnPillowShowcase({ palette }) {
  const p = palette;
  return (
    <div style={{
      background: p.bg,
      height: '100%',
      padding: 40,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: '"Quicksand", system-ui',
    }}>
      <div style={{ position: 'relative', width: 260, height: 220, display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: 24 }}>
        {/* Pillow background */}
        <div style={{ position: 'absolute', zIndex: 1 }}>
          <PillowFlat palette={p} size={240} dimples />
        </div>
        {/* Kek on top, slightly off-center */}
        <div style={{ position: 'absolute', zIndex: 2, top: 20, left: 30 }}>
          <KekBird palette={p} state="sleep" size={140} />
        </div>
      </div>
      <div className="kk-display" style={{ fontSize: 18, fontWeight: 500, color: p.ink, textAlign: 'center', letterSpacing: -0.3 }}>
        Kek on pillow
      </div>
      <div style={{ fontSize: 13, color: p.inkSoft, textAlign: 'center', marginTop: 8 }}>
        The active sleep state — Kek nestles in for the nap
      </div>
    </div>
  );
}

Object.assign(window, { KekBird, KekShowcase, KekStates, IconShowcase, KekOnPillowShowcase });

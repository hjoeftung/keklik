// Phone shell + shared UI primitives for Keklik

function PhoneShell({ palette, children, dark = false, statusDark = null, contentBg = null }) {
  const p = palette;
  const bg = dark ? p.nightSkyDeep : (contentBg || p.bg);
  const sd = statusDark != null ? statusDark : dark;
  return (
    <div className="kk" style={{
      width: 390, height: 780, borderRadius: 44, overflow: 'hidden',
      position: 'relative', background: bg,
      boxShadow: '0 20px 50px rgba(60,40,20,0.18), 0 0 0 1px rgba(0,0,0,0.06)',
    }}>
      {/* status bar */}
      <div style={{
        position: 'absolute', top: 0, left: 0, right: 0, zIndex: 30,
        height: 44, padding: '14px 28px 0', display: 'flex', justifyContent: 'space-between',
        alignItems: 'center', color: sd ? '#fff' : p.ink,
        fontFamily: '"Quicksand", system-ui', fontSize: 15, fontWeight: 600,
      }}>
        <span>9:41</span>
        <div style={{ display: 'flex', gap: 6, alignItems: 'center', opacity: 0.85 }}>
          <svg width="16" height="11" viewBox="0 0 16 11" fill="currentColor"><rect x="0" y="7" width="3" height="4" rx="0.5"/><rect x="4.5" y="5" width="3" height="6" rx="0.5"/><rect x="9" y="2.5" width="3" height="8.5" rx="0.5"/><rect x="13.5" y="0" width="3" height="11" rx="0.5"/></svg>
          <svg width="22" height="11" viewBox="0 0 22 11" fill="none" stroke="currentColor" strokeWidth="1"><rect x="0.5" y="0.5" width="18" height="10" rx="2.5"/><rect x="2" y="2" width="15" height="7" rx="1.2" fill="currentColor"/><rect x="20" y="3.5" width="1.5" height="4" rx="0.5" fill="currentColor"/></svg>
        </div>
      </div>
      {/* dynamic island */}
      <div style={{
        position: 'absolute', top: 11, left: '50%', transform: 'translateX(-50%)',
        width: 110, height: 32, borderRadius: 20, background: '#000', zIndex: 40,
      }} />
      {children}
      {/* home indicator */}
      <div style={{
        position: 'absolute', bottom: 8, left: 0, right: 0, height: 5, zIndex: 50,
        display: 'flex', justifyContent: 'center', pointerEvents: 'none',
      }}>
        <div style={{ width: 134, height: 5, borderRadius: 3, background: dark ? 'rgba(255,255,255,0.6)' : 'rgba(46,42,51,0.35)' }}/>
      </div>
    </div>
  );
}

function TabBar({ palette, active = 'sleep', dark = false }) {
  const p = palette;
  const bg = dark ? 'rgba(27,23,58,0.85)' : '#fff';
  const border = dark ? 'rgba(255,255,255,0.08)' : p.border;
  const muted = dark ? 'rgba(255,255,255,0.5)' : p.inkMuted;
  const tabs = [
    { id: 'sleep', label: 'Sleep', icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
        <path d="M5 9 Q5 5.5, 8 6 Q12 4.5, 16 6 Q19 5.5, 19 9 Q21 12, 19 15 Q19 17, 16 16.5 Q12 18, 8 16.5 Q5 17, 5 15 Q3 12, 5 9 Z" stroke="currentColor" strokeWidth="1.8" fill={active==='sleep' ? 'currentColor' : 'none'} fillOpacity={active==='sleep' ? 0.18 : 0} strokeLinejoin="round"/>
      </svg>
    )},
    { id: 'stats', label: 'Stats', icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round">
        <rect x="4" y="13" width="4" height="7" rx="1.2" fill={active==='stats' ? 'currentColor' : 'none'} fillOpacity={active==='stats'?0.18:0}/>
        <rect x="10" y="9" width="4" height="11" rx="1.2" fill={active==='stats' ? 'currentColor' : 'none'} fillOpacity={active==='stats'?0.18:0}/>
        <rect x="16" y="5" width="4" height="15" rx="1.2" fill={active==='stats' ? 'currentColor' : 'none'} fillOpacity={active==='stats'?0.18:0}/>
      </svg>
    )},
    { id: 'settings', label: 'Settings', icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round">
        <circle cx="12" cy="12" r="3" fill={active==='settings' ? 'currentColor' : 'none'} fillOpacity={active==='settings'?0.18:0}/>
        <path d="M12 3v2M12 19v2M3 12h2M19 12h2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M5.6 18.4L7 17M17 7l1.4-1.4"/>
      </svg>
    )},
  ];
  return (
    <div style={{
      position: 'absolute', bottom: 0, left: 0, right: 0, zIndex: 20,
      paddingBottom: 22, paddingTop: 10,
      background: bg, borderTop: `1px solid ${border}`,
      display: 'flex', justifyContent: 'space-around',
    }}>
      {tabs.map(t => (
        <div key={t.id} style={{
          display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3,
          color: t.id === active ? p.primary : muted, padding: '4px 12px',
        }}>
          {t.icon}
          <span style={{ fontSize: 11, fontWeight: 600, letterSpacing: 0.2 }}>{t.label}</span>
        </div>
      ))}
    </div>
  );
}

function Card({ palette, children, style = {}, soft = false }) {
  const p = palette;
  return (
    <div style={{
      background: soft ? p.surfaceSoft : p.surface,
      borderRadius: 24,
      padding: 18,
      ...style,
    }}>{children}</div>
  );
}

function Pill({ palette, children, active = false, style = {} }) {
  const p = palette;
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 6,
      padding: '6px 12px', borderRadius: 999,
      background: active ? p.ink : p.surface,
      color: active ? p.surface : p.inkSoft,
      fontSize: 13, fontWeight: 600,
      border: active ? 'none' : `1px solid ${p.border}`,
      ...style,
    }}>{children}</span>
  );
}

// Big rounded button — used in forms, dialogs
function BigButton({ palette, children, primary = true, style = {} }) {
  const p = palette;
  return (
    <button style={{
      display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
      width: '100%', padding: '18px 24px', borderRadius: 22, border: 'none',
      background: primary ? p.primary : p.surfaceSoft,
      color: primary ? '#fff' : p.ink,
      fontFamily: 'inherit', fontSize: 17, fontWeight: 600,
      cursor: 'pointer', boxShadow: primary ? `0 6px 18px ${p.primary}40` : 'none',
      ...style,
    }}>{children}</button>
  );
}

// Section header used in stats etc
function SectionLabel({ palette, children, style = {} }) {
  const p = palette;
  return (
    <div style={{
      fontSize: 12, fontWeight: 700, color: p.inkMuted, letterSpacing: 0.6,
      textTransform: 'uppercase', ...style,
    }}>{children}</div>
  );
}

Object.assign(window, { PhoneShell, TabBar, Card, Pill, BigButton, SectionLabel });

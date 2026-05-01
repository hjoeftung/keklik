// Sleep screen variations — different pillow button takes and active states.
// Each is a self-contained mobile screen rendered inside <PhoneShell>.

// Top header used across screens
function SleepHeader({ palette, baby = 'Sofia', mode = 'awake', dark = false }) {
  const p = palette;
  const fg = dark ? '#fff' : p.ink;
  const sub = dark ? 'rgba(255,255,255,0.65)' : p.inkSoft;
  // Date string varies subtly by mode so the header still anchors context
  const date = 'Tuesday · April 28';
  return (
    <div style={{ paddingTop: 56, padding: '56px 24px 0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <div>
        <div style={{ fontSize: 22, fontWeight: 600, color: fg }} className="kk-display">{baby}</div>
        <div style={{ fontSize: 13, fontWeight: 500, color: sub, marginTop: 2 }}>{date}</div>
      </div>
      <div style={{
        width: 40, height: 40, borderRadius: 20,
        background: dark ? 'rgba(255,255,255,0.1)' : p.surface,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        border: dark ? '1px solid rgba(255,255,255,0.12)' : `1px solid ${p.border}`,
        fontWeight: 700, color: dark ? '#fff' : p.primary, fontSize: 14,
      }}>S</div>
    </div>
  );
}

// ─── Variation A: Big literal pillow button, awake state ────────────────
function SleepA_Awake({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p}>
      <SleepHeader palette={p} mode="awake" />
      <div style={{ position: 'absolute', top: 130, left: 0, right: 0, padding: '0 24px', textAlign: 'center' }}>
        <div style={{ fontSize: 13, fontWeight: 600, color: p.inkMuted, letterSpacing: 0.4, textTransform: 'uppercase' }}>Active for</div>
        <div className="kk-display kk-num" style={{ fontSize: 52, fontWeight: 500, color: p.ink, lineHeight: 1, marginTop: 6, letterSpacing: -1 }}>
          2<span style={{ fontSize: 28, color: p.inkSoft, fontWeight: 500 }}>h </span>14<span style={{ fontSize: 28, color: p.inkSoft, fontWeight: 500 }}>m</span>
        </div>
      </div>
      {/* Pillow button */}
      <div style={{ position: 'absolute', top: 290, left: '50%', transform: 'translateX(-50%)', textAlign: 'center' }}>
        <div style={{
          padding: 0, background: 'transparent', border: 'none',
          filter: `drop-shadow(0 16px 32px ${p.primary}40)`,
          animation: 'kk-breathe 4s ease-in-out infinite',
        }}>
          <PillowFlat palette={p} size={260} label="Tap to sleep" />
        </div>
        <div style={{ marginTop: 12, color: p.inkSoft, fontSize: 14, fontWeight: 500 }}>
          Last sleep ended 14:32
        </div>
      </div>
      {/* Quick actions */}
      <div style={{ position: 'absolute', bottom: 110, left: 24, right: 24, display: 'flex', gap: 12 }}>
        <Card palette={p} soft style={{ flex: 1, padding: 14, display: 'flex', alignItems: 'center', gap: 10 }}>
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke={p.primaryDeep} strokeWidth="1.8" strokeLinecap="round">
            <circle cx="10" cy="10" r="7"/><path d="M10 6v4l2.5 2"/>
          </svg>
          <div>
            <div style={{ fontSize: 12, color: p.inkMuted, fontWeight: 600 }}>Log past</div>
            <div style={{ fontSize: 14, color: p.ink, fontWeight: 600 }}>sleep</div>
          </div>
        </Card>
        <Card palette={p} soft style={{ flex: 1, padding: 14, display: 'flex', alignItems: 'center', gap: 10 }}>
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke={p.primaryDeep} strokeWidth="1.8" strokeLinecap="round">
            <rect x="3" y="4" width="14" height="13" rx="2"/><path d="M7 2v3M13 2v3M3 8h14"/>
          </svg>
          <div>
            <div style={{ fontSize: 12, color: p.inkMuted, fontWeight: 600 }}>Today</div>
            <div style={{ fontSize: 14, color: p.ink, fontWeight: 600 }}>4h 22m sleep</div>
          </div>
        </Card>
      </div>
      <TabBar palette={p} active="sleep" />
    </PhoneShell>
  );
}

// ─── Variation B: Pillow button, asleep state — soft daytime palette stays
function SleepB_Asleep_Soft({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p}>
      <SleepHeader palette={p} mode="asleep" />
      <div style={{ position: 'absolute', top: 130, left: 0, right: 0, padding: '0 24px', textAlign: 'center' }}>
        <div style={{ fontSize: 13, fontWeight: 600, color: p.inkMuted, letterSpacing: 0.4, textTransform: 'uppercase' }}>Sleeping for</div>
        <div className="kk-display kk-num" style={{ fontSize: 52, fontWeight: 500, color: p.ink, lineHeight: 1, marginTop: 6, letterSpacing: -1 }}>
          0<span style={{ fontSize: 28, color: p.inkSoft, fontWeight: 500 }}>h </span>47<span style={{ fontSize: 28, color: p.inkSoft, fontWeight: 500 }}>m</span>
        </div>
        <div style={{ marginTop: 6, fontSize: 13, color: p.inkSoft }}>Started at 14:32 · <span style={{ color: p.primary, fontWeight: 600 }}>Edit</span></div>
      </div>
      {/* Pillow with mask */}
      <div style={{ position: 'absolute', top: 280, left: '50%', transform: 'translateX(-50%)', textAlign: 'center' }}>
        <div style={{
          filter: `drop-shadow(0 16px 32px ${p.night}30)`,
          animation: 'kk-breathe 4s ease-in-out infinite',
          position: 'relative',
        }}>
          <PillowMasked palette={p} size={260} />
          {/* floating Z's */}
          <div style={{ position: 'absolute', top: -10, right: 30, fontSize: 28, fontWeight: 700, color: p.night, fontFamily: 'Quicksand', opacity: 0.7 }}>z</div>
          <div style={{ position: 'absolute', top: -28, right: 50, fontSize: 22, fontWeight: 700, color: p.night, fontFamily: 'Quicksand', opacity: 0.5 }}>z</div>
          <div style={{ position: 'absolute', top: -42, right: 68, fontSize: 16, fontWeight: 700, color: p.night, fontFamily: 'Quicksand', opacity: 0.35 }}>z</div>
        </div>
        <div style={{ marginTop: 24, fontSize: 14, fontWeight: 600, color: p.night }}>
          Tap pillow to wake · shhh
        </div>
      </div>
      <TabBar palette={p} active="sleep" />
    </PhoneShell>
  );
}

// ─── Variation C: Asleep — full dark night-sky takeover
function SleepC_Asleep_Night({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p} dark statusDark>
      {/* night sky background */}
      <div style={{ position: 'absolute', inset: 0, zIndex: 0 }}>
        <NightScene width={390} height={780} palette={p} />
      </div>
      <div style={{ position: 'relative', zIndex: 1 }}>
        <SleepHeader palette={p} mode="asleep" dark />
        <div style={{ padding: '40px 24px 0', textAlign: 'center' }}>
          <div style={{ fontSize: 13, fontWeight: 600, color: 'rgba(255,255,255,0.6)', letterSpacing: 0.4, textTransform: 'uppercase' }}>Sleeping for</div>
          <div className="kk-display kk-num" style={{ fontSize: 56, fontWeight: 500, color: '#fff', lineHeight: 1, marginTop: 6, letterSpacing: -1 }}>
            1<span style={{ fontSize: 30, color: 'rgba(255,255,255,0.6)' }}>h </span>23<span style={{ fontSize: 30, color: 'rgba(255,255,255,0.6)' }}>m</span>
          </div>
          <div style={{ marginTop: 6, fontSize: 13, color: 'rgba(255,255,255,0.55)' }}>Started at 19:47</div>
        </div>
        {/* Pillow */}
        <div style={{ marginTop: 36, textAlign: 'center', position: 'relative' }}>
          <div style={{
            display: 'inline-block', filter: `drop-shadow(0 16px 32px rgba(0,0,0,0.5))`,
            animation: 'kk-breathe 4.5s ease-in-out infinite',
          }}>
            <PillowFlat palette={p} size={240} label="Tap to wake" />
          </div>
          {/* floating Zs */}
          <div style={{ position: 'absolute', top: -12, right: 60, fontSize: 26, fontWeight: 700, color: p.moon, opacity: 0.7 }}>z</div>
          <div style={{ position: 'absolute', top: -32, right: 78, fontSize: 18, fontWeight: 700, color: p.moon, opacity: 0.5 }}>z</div>
        </div>
        <div style={{ marginTop: 22, textAlign: 'center' }}>
          <span style={{
            display: 'inline-flex', alignItems: 'center', gap: 8,
            padding: '10px 18px', borderRadius: 999,
            background: 'rgba(255,255,255,0.12)',
            border: '1px solid rgba(255,255,255,0.22)',
            color: '#fff', fontSize: 14, fontWeight: 600,
            backdropFilter: 'blur(8px)',
          }}>
            <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M3 11l1.5-4.5L9 2l3 3-4.5 4.5L3 11zM2 13h12"/></svg>
            Edit start time · 19:47
          </span>
        </div>
      </div>
      <TabBar palette={p} active="sleep" dark />
    </PhoneShell>
  );
}

// ─── Variation D: Stylized pillow — squircle with corner dots ────────────
function SleepD_Stylized({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p}>
      <SleepHeader palette={p} mode="awake" />
      <div style={{ position: 'absolute', top: 130, left: 0, right: 0, padding: '0 24px', textAlign: 'center' }}>
        <div style={{ fontSize: 13, fontWeight: 600, color: p.inkMuted, letterSpacing: 0.4, textTransform: 'uppercase' }}>Active for</div>
        <div className="kk-display kk-num" style={{ fontSize: 52, fontWeight: 500, color: p.ink, lineHeight: 1, marginTop: 6, letterSpacing: -1 }}>
          2<span style={{ fontSize: 28, color: p.inkSoft }}>h </span>14<span style={{ fontSize: 28, color: p.inkSoft }}>m</span>
        </div>
      </div>
      {/* Stylized pillow — squircle */}
      <div style={{ position: 'absolute', top: 290, left: '50%', transform: 'translateX(-50%)', textAlign: 'center' }}>
        <div style={{
          width: 240, height: 200, borderRadius: 64, position: 'relative',
          background: p.primary, boxShadow: `0 24px 50px ${p.primary}55, inset 0 -8px 16px rgba(0,0,0,0.06)`,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          animation: 'kk-breathe 4s ease-in-out infinite',
        }}>
          {/* inner highlight */}
          <div style={{ position: 'absolute', top: 12, left: 24, right: 24, height: 60, borderRadius: 40, background: 'rgba(255,255,255,0.25)', filter: 'blur(8px)' }}/>
          {/* corner dimples */}
          <div style={{ position: 'absolute', top: 14, left: 14, width: 6, height: 6, borderRadius: 3, background: 'rgba(0,0,0,0.18)' }}/>
          <div style={{ position: 'absolute', top: 14, right: 14, width: 6, height: 6, borderRadius: 3, background: 'rgba(0,0,0,0.18)' }}/>
          <div style={{ position: 'absolute', bottom: 14, left: 14, width: 6, height: 6, borderRadius: 3, background: 'rgba(0,0,0,0.18)' }}/>
          <div style={{ position: 'absolute', bottom: 14, right: 14, width: 6, height: 6, borderRadius: 3, background: 'rgba(0,0,0,0.18)' }}/>
          <span style={{ color: '#fff', fontSize: 22, fontWeight: 600, letterSpacing: 0.2, position: 'relative' }}>Tap to sleep</span>
        </div>
      </div>
      <TabBar palette={p} active="sleep" />
    </PhoneShell>
  );
}

// ─── Variation E: Subtle / minimal — just color shift, breathing pillow
function SleepE_Minimal({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p} contentBg={p.bg}>
      <div style={{ paddingTop: 56, padding: '56px 24px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <div style={{ fontSize: 13, fontWeight: 600, color: p.inkSoft }}>Tuesday · April 27</div>
          <div className="kk-display" style={{ fontSize: 28, fontWeight: 500, color: p.ink, marginTop: 4, letterSpacing: -0.5 }}>Hi, Sofia</div>
        </div>
      </div>
      <div style={{ position: 'absolute', inset: 0, top: 200, display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '0 24px' }}>
        <div style={{ fontSize: 13, color: p.inkMuted, fontWeight: 600, letterSpacing: 0.4, textTransform: 'uppercase', marginBottom: 8 }}>Active for</div>
        <div className="kk-display kk-num" style={{ fontSize: 64, fontWeight: 500, color: p.ink, lineHeight: 1, letterSpacing: -1.5 }}>
          2:14
        </div>
        <div style={{ fontSize: 13, color: p.inkSoft, marginTop: 6 }}>Last slept 12:18 — 14:32</div>
        <div style={{ marginTop: 40, animation: 'kk-breathe 4s ease-in-out infinite' }}>
          <PillowFlat palette={p} size={220} label="Sleep" />
        </div>
        <button style={{
          marginTop: 24, padding: '10px 18px', borderRadius: 999, border: `1px solid ${p.border}`,
          background: 'transparent', color: p.inkSoft, fontFamily: 'inherit', fontSize: 14, fontWeight: 600,
        }}>+ Log past sleep</button>
      </div>
      <TabBar palette={p} active="sleep" />
    </PhoneShell>
  );
}

// ─── Variation F: Empty state, brand new family ────────────────────────
function SleepF_Empty({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p}>
      <SleepHeader palette={p} mode="other" />
      <div style={{ position: 'absolute', inset: 0, top: 150, display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '0 32px', textAlign: 'center' }}>
        <Bun palette={p} size={120} />
        <div className="kk-display" style={{ fontSize: 24, fontWeight: 500, color: p.ink, marginTop: 16, letterSpacing: -0.4 }}>
          Sofia hasn't slept yet
        </div>
        <div style={{ fontSize: 14, color: p.inkSoft, marginTop: 8, lineHeight: 1.5, maxWidth: 280 }}>
          Tap the pillow when she dozes off. We'll keep track from there.
        </div>
        <div style={{ marginTop: 32, animation: 'kk-breathe 4s ease-in-out infinite', filter: `drop-shadow(0 12px 24px ${p.primary}50)` }}>
          <PillowFlat palette={p} size={200} label="Start sleep" />
        </div>
        <button style={{
          marginTop: 28, padding: '10px 18px', borderRadius: 999, border: 'none',
          background: 'transparent', color: p.primary, fontFamily: 'inherit', fontSize: 14, fontWeight: 600,
        }}>or log a past sleep →</button>
      </div>
      <TabBar palette={p} active="sleep" />
    </PhoneShell>
  );
}

Object.assign(window, { SleepA_Awake, SleepB_Asleep_Soft, SleepC_Asleep_Night, SleepD_Stylized, SleepE_Minimal, SleepF_Empty });

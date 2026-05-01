// Onboarding, Settings, Forms (log past session, edit time)

// ─── Onboarding 1: Welcome ─────────────────────────────────────────────
function Onboard1({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p}>
      <div style={{ position: 'absolute', top: 100, left: 0, right: 0, padding: '0 32px', textAlign: 'center' }}>
        {/* hero */}
        <div style={{ position: 'relative', height: 220, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div style={{ position: 'absolute', top: 0, left: 30, animation: 'kk-float 3.5s ease-in-out infinite' }}>
            <Cloud size={80} palette={p} />
          </div>
          <div style={{ position: 'absolute', top: 30, right: 36 }}>
            <Moon size={52} palette={p} />
          </div>
          <div style={{ position: 'absolute', top: 14, right: 110 }}>
            <Star size={14} palette={p} twinkle delay={0} />
          </div>
          <div style={{ position: 'absolute', top: 60, left: 100 }}>
            <Star size={10} palette={p} twinkle delay={0.6} />
          </div>
          <div style={{ filter: `drop-shadow(0 14px 28px ${p.primary}40)` }}>
            <PillowFlat palette={p} size={200} />
          </div>
        </div>
        <div className="kk-display" style={{ fontSize: 32, fontWeight: 500, color: p.ink, marginTop: 30, letterSpacing: -0.6, lineHeight: 1.15 }}>
          A cozy little<br/>sleep tracker
        </div>
        <div style={{ fontSize: 15, color: p.inkSoft, marginTop: 14, lineHeight: 1.5 }}>
          Tap a pillow when baby naps.<br/>Watch the day fill up with rest.
        </div>
      </div>
      <div style={{ position: 'absolute', bottom: 36, left: 24, right: 24 }}>
        <BigButton palette={p}>Continue with Google</BigButton>
        <div style={{ display: 'flex', justifyContent: 'center', gap: 6, marginTop: 18 }}>
          <span style={{ width: 24, height: 6, borderRadius: 3, background: p.primary }}/>
          <span style={{ width: 6, height: 6, borderRadius: 3, background: p.border }}/>
        </div>
      </div>
    </PhoneShell>
  );
}

// ─── Onboarding 2: Set up baby ─────────────────────────────────────────
function Onboard2({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p}>
      <div style={{ paddingTop: 70, padding: '70px 24px 0' }}>
        <div style={{ fontSize: 13, fontWeight: 600, color: p.primary, letterSpacing: 0.4, textTransform: 'uppercase' }}>Step 2 of 2</div>
        <div className="kk-display" style={{ fontSize: 28, fontWeight: 500, color: p.ink, marginTop: 6, letterSpacing: -0.5, lineHeight: 1.2 }}>
          Tell us about your<br/>little one
        </div>
        <div style={{ marginTop: 30, display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div>
            <div style={{ fontSize: 12, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase', marginBottom: 8 }}>Baby's name</div>
            <div style={{
              padding: '16px 18px', background: p.surface, borderRadius: 18,
              border: `1.5px solid ${p.primary}`, fontSize: 17, color: p.ink, fontWeight: 500,
              display: 'flex', alignItems: 'center', gap: 8,
            }}>
              Sofia<span style={{ width: 1.5, height: 18, background: p.primary, animation: 'kk-twinkle 1s steps(2) infinite' }}/>
            </div>
          </div>
          <div>
            <div style={{ fontSize: 12, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase', marginBottom: 8 }}>Birthday</div>
            <div style={{
              padding: '16px 18px', background: p.surface, borderRadius: 18,
              border: `1px solid ${p.border}`, fontSize: 17, color: p.ink, fontWeight: 500,
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            }}>
              <span>March 14, 2025</span>
              <span style={{ fontSize: 13, color: p.primary, fontWeight: 600 }}>Edit</span>
            </div>
          </div>
          <div>
            <div style={{ fontSize: 12, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase', marginBottom: 8 }}>Night window</div>
            <div style={{
              padding: '14px 18px', background: p.surface, borderRadius: 18,
              border: `1px solid ${p.border}`, display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <Moon size={28} palette={p} />
                <div>
                  <div className="kk-num" style={{ fontSize: 17, fontWeight: 600, color: p.ink }}>19:00 – 08:00</div>
                  <div style={{ fontSize: 12, color: p.inkSoft }}>Used to classify naps vs night sleep</div>
                </div>
              </div>
            </div>
          </div>
          {/* tiny illustration */}
          <div style={{ display: 'flex', justifyContent: 'center', marginTop: 12 }}>
            <Bun palette={p} size={80} />
          </div>
        </div>
      </div>
      <div style={{ position: 'absolute', bottom: 36, left: 24, right: 24 }}>
        <BigButton palette={p}>Start tracking</BigButton>
        <div style={{ display: 'flex', justifyContent: 'center', gap: 6, marginTop: 18 }}>
          <span style={{ width: 6, height: 6, borderRadius: 3, background: p.border }}/>
          <span style={{ width: 24, height: 6, borderRadius: 3, background: p.primary }}/>
        </div>
      </div>
    </PhoneShell>
  );
}

// ─── Settings ─────────────────────────────────────────────────────────
function Settings({ palette }) {
  const p = palette;
  const Row = ({ icon, label, value, last }) => (
    <div style={{
      display: 'flex', alignItems: 'center', padding: '14px 18px',
      borderBottom: last ? 'none' : `1px solid ${p.border}`,
    }}>
      <div style={{ width: 36, height: 36, borderRadius: 12, background: p.primarySoft, display: 'flex', alignItems: 'center', justifyContent: 'center', marginRight: 12 }}>{icon}</div>
      <div style={{ flex: 1, fontSize: 15, color: p.ink, fontWeight: 600 }}>{label}</div>
      <div style={{ fontSize: 14, color: p.inkSoft, fontWeight: 500 }}>{value}</div>
      <svg width="8" height="14" viewBox="0 0 8 14" style={{ marginLeft: 10 }}><path d="M1 1l6 6-6 6" stroke={p.inkMuted} strokeWidth="1.8" fill="none" strokeLinecap="round"/></svg>
    </div>
  );
  return (
    <PhoneShell palette={p}>
      <div style={{ paddingTop: 56, padding: '56px 24px 12px' }}>
        <div className="kk-display" style={{ fontSize: 28, fontWeight: 500, color: p.ink, letterSpacing: -0.5 }}>Settings</div>
      </div>
      <div className="kk-scroll" style={{ position: 'absolute', top: 110, bottom: 80, left: 0, right: 0, padding: '0 16px' }}>
        {/* Family card */}
        <div style={{
          background: p.primarySoft, borderRadius: 24, padding: 18, marginBottom: 18,
          display: 'flex', alignItems: 'center', gap: 14, position: 'relative', overflow: 'hidden',
        }}>
          <Bun palette={p} size={64} />
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 12, fontWeight: 700, color: p.primaryDeep, letterSpacing: 0.4, textTransform: 'uppercase' }}>Baby</div>
            <div className="kk-display" style={{ fontSize: 22, fontWeight: 500, color: p.ink, letterSpacing: -0.3 }}>Sofia</div>
            <div style={{ fontSize: 13, color: p.inkSoft, marginTop: 2 }}>13 months · 2 caregivers</div>
          </div>
          <div style={{ width: 32, height: 32, borderRadius: 16, background: 'rgba(255,255,255,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke={p.primaryDeep} strokeWidth="2" strokeLinecap="round"><path d="M3 11l1.5-4.5L9 2l3 3-4.5 4.5L3 11zM2 12h10"/></svg>
          </div>
        </div>

        <SectionLabel palette={p} style={{ padding: '0 8px 8px' }}>Sleep</SectionLabel>
        <div style={{ background: p.surface, borderRadius: 22, marginBottom: 18, overflow: 'hidden' }}>
          <Row icon={<Moon size={20} palette={p} />} label="Night window" value="19:00 – 08:00" />
          <Row icon={<svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke={p.primaryDeep} strokeWidth="1.8" strokeLinecap="round"><circle cx="10" cy="10" r="7"/><path d="M10 6v4l2.5 2"/></svg>} label="Day starts at" value="06:00" />
          <Row icon={<svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke={p.primaryDeep} strokeWidth="1.8" strokeLinecap="round"><path d="M3 10h14M3 6h14M3 14h14"/></svg>} label="Time format" value="24-hour" last />
        </div>

        <SectionLabel palette={p} style={{ padding: '0 8px 8px' }}>Family</SectionLabel>
        <div style={{ background: p.surface, borderRadius: 22, marginBottom: 18, overflow: 'hidden' }}>
          <Row icon={<svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke={p.primaryDeep} strokeWidth="1.8" strokeLinecap="round"><circle cx="10" cy="7" r="3"/><path d="M3 17c0-3.5 3-6 7-6s7 2.5 7 6"/></svg>} label="Caregivers" value="2" />
          <Row icon={<svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke={p.primaryDeep} strokeWidth="1.8" strokeLinecap="round"><path d="M3 7l7 5 7-5M3 7v8a1 1 0 001 1h12a1 1 0 001-1V7M3 7l1-2h12l1 2"/></svg>} label="Invite a partner" value="" last />
        </div>

        <SectionLabel palette={p} style={{ padding: '0 8px 8px' }}>About</SectionLabel>
        <div style={{ background: p.surface, borderRadius: 22, overflow: 'hidden' }}>
          <Row icon={<svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke={p.primaryDeep} strokeWidth="1.8" strokeLinecap="round"><circle cx="10" cy="10" r="7"/><path d="M10 6v5M10 14h.01"/></svg>} label="Help & feedback" value="" />
          <Row icon={<svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke={p.primaryDeep} strokeWidth="1.8" strokeLinecap="round"><path d="M7 4h10v12H7zM3 8v8a2 2 0 002 2h10"/></svg>} label="Sign out" value="" last />
        </div>
        <div style={{ textAlign: 'center', padding: '24px 0 8px', fontSize: 12, color: p.inkMuted }}>
          Keklik · v0.4.0 · made with ♡
        </div>
      </div>
      <TabBar palette={p} active="settings" />
    </PhoneShell>
  );
}

// ─── Log past session form ─────────────────────────────────────────────
function LogPast({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p} contentBg={p.bg}>
      {/* dim background hint */}
      <div style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.25)' }}/>
      {/* sheet */}
      <div style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        background: p.bg, borderTopLeftRadius: 28, borderTopRightRadius: 28,
        padding: '12px 24px 32px', boxShadow: '0 -10px 30px rgba(0,0,0,0.18)',
      }}>
        <div style={{ width: 44, height: 5, borderRadius: 3, background: p.border, margin: '0 auto 16px' }}/>
        <div className="kk-display" style={{ fontSize: 22, fontWeight: 500, color: p.ink, letterSpacing: -0.3 }}>Log past sleep</div>
        <div style={{ fontSize: 13, color: p.inkSoft, marginTop: 4 }}>For sleep that wasn't tracked live</div>

        {/* Type toggle */}
        <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
          <div style={{
            flex: 1, padding: '12px 14px', borderRadius: 16,
            background: p.night, color: '#fff', display: 'flex', alignItems: 'center', gap: 8,
            fontWeight: 600, fontSize: 14,
          }}>
            <Moon size={20} palette={p} mode="crescent" />Night sleep
          </div>
          <div style={{
            flex: 1, padding: '12px 14px', borderRadius: 16,
            background: p.surface, color: p.inkSoft, border: `1px solid ${p.border}`,
            display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600, fontSize: 14,
          }}>
            <svg width="20" height="20" viewBox="0 0 20 20" fill={p.nap}><circle cx="10" cy="10" r="4"/><g stroke={p.nap} strokeWidth="1.8" strokeLinecap="round"><path d="M10 2v2M10 16v2M2 10h2M16 10h2M4 4l1.5 1.5M14.5 14.5L16 16M4 16l1.5-1.5M14.5 5.5L16 4"/></g></svg>
            Nap
          </div>
        </div>

        {/* Time pickers */}
        <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ background: p.surface, borderRadius: 18, padding: '14px 18px', border: `1px solid ${p.border}`, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <div style={{ fontSize: 12, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase' }}>Started</div>
              <div className="kk-num" style={{ fontSize: 22, fontWeight: 600, color: p.ink, marginTop: 2 }}>Today, 13:00</div>
            </div>
            <div style={{ fontSize: 13, color: p.primary, fontWeight: 600 }}>Change</div>
          </div>
          <div style={{ background: p.surface, borderRadius: 18, padding: '14px 18px', border: `1px solid ${p.border}`, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <div style={{ fontSize: 12, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase' }}>Ended</div>
              <div className="kk-num" style={{ fontSize: 22, fontWeight: 600, color: p.ink, marginTop: 2 }}>Today, 14:30</div>
            </div>
            <div style={{ fontSize: 13, color: p.primary, fontWeight: 600 }}>Change</div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '0 4px', fontSize: 13, color: p.inkSoft }}>
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={p.success} strokeWidth="2" strokeLinecap="round"><path d="M3 8l3 3 7-7"/></svg>
            Duration <span className="kk-num" style={{ color: p.ink, fontWeight: 600 }}>1h 30m</span>
          </div>
        </div>

        <div style={{ display: 'flex', gap: 10, marginTop: 22 }}>
          <BigButton palette={p} primary={false} style={{ flex: 1 }}>Cancel</BigButton>
          <BigButton palette={p} style={{ flex: 1.4 }}>Save sleep</BigButton>
        </div>
      </div>
    </PhoneShell>
  );
}

// ─── Log past sleep — default state (just opened, suggested defaults) ───
function LogPastDefault({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p} contentBg={p.bg}>
      <div style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.25)' }}/>
      <div style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        background: p.bg, borderTopLeftRadius: 28, borderTopRightRadius: 28,
        padding: '12px 24px 32px', boxShadow: '0 -10px 30px rgba(0,0,0,0.18)',
      }}>
        <div style={{ width: 44, height: 5, borderRadius: 3, background: p.border, margin: '0 auto 16px' }}/>
        <div className="kk-display" style={{ fontSize: 22, fontWeight: 500, color: p.ink, letterSpacing: -0.3 }}>Log past sleep</div>
        <div style={{ fontSize: 13, color: p.inkSoft, marginTop: 4 }}>We've suggested times — adjust if needed</div>

        {/* Type defaults to nap (more common to log retroactively) */}
        <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
          <div style={{
            flex: 1, padding: '12px 14px', borderRadius: 16,
            background: p.surface, color: p.inkSoft, border: `1px solid ${p.border}`,
            display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600, fontSize: 14,
          }}>
            <Moon size={20} palette={p} mode="crescent" />Night sleep
          </div>
          <div style={{
            flex: 1, padding: '12px 14px', borderRadius: 16,
            background: p.nap, color: '#fff',
            display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600, fontSize: 14,
          }}>
            <svg width="20" height="20" viewBox="0 0 20 20" fill="#fff"><circle cx="10" cy="10" r="4"/><g stroke="#fff" strokeWidth="1.8" strokeLinecap="round"><path d="M10 2v2M10 16v2M2 10h2M16 10h2M4 4l1.5 1.5M14.5 14.5L16 16M4 16l1.5-1.5M14.5 5.5L16 4"/></g></svg>
            Nap
          </div>
        </div>

        {/* Suggested defaults: end = now, start = 1h ago */}
        <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ background: p.primarySoft, borderRadius: 18, padding: '14px 18px', border: `1.5px dashed ${p.primary}80`, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <div style={{ fontSize: 12, fontWeight: 700, color: p.primaryDeep, letterSpacing: 0.4, textTransform: 'uppercase' }}>Started · suggested</div>
              <div className="kk-num" style={{ fontSize: 22, fontWeight: 600, color: p.ink, marginTop: 2 }}>Today, 14:14</div>
            </div>
            <div style={{ fontSize: 13, color: p.primary, fontWeight: 600 }}>Tap to set</div>
          </div>
          <div style={{ background: p.primarySoft, borderRadius: 18, padding: '14px 18px', border: `1.5px dashed ${p.primary}80`, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <div style={{ fontSize: 12, fontWeight: 700, color: p.primaryDeep, letterSpacing: 0.4, textTransform: 'uppercase' }}>Ended · now</div>
              <div className="kk-num" style={{ fontSize: 22, fontWeight: 600, color: p.ink, marginTop: 2 }}>Today, 15:14</div>
            </div>
            <div style={{ fontSize: 13, color: p.primary, fontWeight: 600 }}>Tap to set</div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '0 4px', fontSize: 13, color: p.inkSoft }}>
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={p.inkMuted} strokeWidth="2" strokeLinecap="round"><circle cx="8" cy="8" r="6.5"/><path d="M8 4v4l2.5 2"/></svg>
            Duration <span className="kk-num" style={{ color: p.ink, fontWeight: 600 }}>1h 0m</span>
          </div>
        </div>

        <div style={{ display: 'flex', gap: 10, marginTop: 22 }}>
          <BigButton palette={p} primary={false} style={{ flex: 1 }}>Cancel</BigButton>
          <BigButton palette={p} style={{ flex: 1.4 }}>Save sleep</BigButton>
        </div>
      </div>
    </PhoneShell>
  );
}

// ─── Edit session detail panel (from stats today) ──────────────────────
function EditSession({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p} contentBg={p.bg}>
      <div style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.3)' }}/>
      <div style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        background: p.bg, borderTopLeftRadius: 28, borderTopRightRadius: 28,
        padding: '12px 24px 32px',
      }}>
        <div style={{ width: 44, height: 5, borderRadius: 3, background: p.border, margin: '0 auto 16px' }}/>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{ width: 44, height: 44, borderRadius: 14, background: p.night, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <Moon size={26} palette={p} />
          </div>
          <div style={{ flex: 1 }}>
            <div className="kk-display" style={{ fontSize: 20, fontWeight: 500, color: p.ink, letterSpacing: -0.3 }}>Night sleep</div>
            <div style={{ fontSize: 13, color: p.inkSoft }} className="kk-num">19:47 – 01:30 · 5h 43m</div>
          </div>
        </div>

        <div style={{ marginTop: 20, display: 'flex', gap: 10 }}>
          <div style={{ flex: 1, background: p.surface, borderRadius: 18, padding: '12px 14px', border: `1px solid ${p.border}` }}>
            <div style={{ fontSize: 11, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase' }}>Start</div>
            <div className="kk-num" style={{ fontSize: 19, fontWeight: 600, color: p.ink, marginTop: 2 }}>19:47</div>
          </div>
          <div style={{ flex: 1, background: p.surface, borderRadius: 18, padding: '12px 14px', border: `1px solid ${p.border}` }}>
            <div style={{ fontSize: 11, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase' }}>End</div>
            <div className="kk-num" style={{ fontSize: 19, fontWeight: 600, color: p.ink, marginTop: 2 }}>01:30</div>
          </div>
        </div>

        <div style={{ marginTop: 14, padding: 14, background: p.napSoft, borderRadius: 16, fontSize: 13, color: p.ink, lineHeight: 1.5 }}>
          <span style={{ fontWeight: 700 }}>Classification</span> · This sleep started in the night window (19:00–08:00), so it's counted as night sleep.
        </div>

        <div style={{ display: 'flex', gap: 10, marginTop: 18 }}>
          <BigButton palette={p} primary={false} style={{ flex: 1, color: p.danger }}>Delete</BigButton>
          <BigButton palette={p} style={{ flex: 1.4 }}>Save changes</BigButton>
        </div>
      </div>
    </PhoneShell>
  );
}

// ─── Log past — empty/initial state ───────────────────────────────────
function LogPastEmpty({ palette }) {
  const p = palette;
  return (
    <PhoneShell palette={p} contentBg={p.bg}>
      <div style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.25)' }}/>
      <div style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        background: p.bg, borderTopLeftRadius: 28, borderTopRightRadius: 28,
        padding: '12px 24px 32px', boxShadow: '0 -10px 30px rgba(0,0,0,0.18)',
      }}>
        <div style={{ width: 44, height: 5, borderRadius: 3, background: p.border, margin: '0 auto 16px' }}/>
        <div className="kk-display" style={{ fontSize: 22, fontWeight: 500, color: p.ink, letterSpacing: -0.3 }}>Log past sleep</div>
        <div style={{ fontSize: 13, color: p.inkSoft, marginTop: 4 }}>For sleep that wasn't tracked live</div>

        {/* Type toggle — neither selected yet */}
        <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
          <div style={{
            flex: 1, padding: '12px 14px', borderRadius: 16,
            background: p.surface, color: p.inkSoft, border: `1px dashed ${p.border}`,
            display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600, fontSize: 14,
          }}>
            <Moon size={20} palette={p} mode="crescent" />Night sleep
          </div>
          <div style={{
            flex: 1, padding: '12px 14px', borderRadius: 16,
            background: p.surface, color: p.inkSoft, border: `1px dashed ${p.border}`,
            display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600, fontSize: 14,
          }}>
            <svg width="20" height="20" viewBox="0 0 20 20" fill={p.nap}><circle cx="10" cy="10" r="4"/><g stroke={p.nap} strokeWidth="1.8" strokeLinecap="round"><path d="M10 2v2M10 16v2M2 10h2M16 10h2M4 4l1.5 1.5M14.5 14.5L16 16M4 16l1.5-1.5M14.5 5.5L16 4"/></g></svg>
            Nap
          </div>
        </div>

        {/* Empty time pickers — placeholders */}
        <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ background: p.surface, borderRadius: 18, padding: '14px 18px', border: `1px dashed ${p.border}`, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <div style={{ fontSize: 12, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase' }}>Started</div>
              <div className="kk-num" style={{ fontSize: 22, fontWeight: 500, color: p.inkMuted, marginTop: 2 }}>Pick a time</div>
            </div>
            <div style={{ fontSize: 13, color: p.primary, fontWeight: 600 }}>Set</div>
          </div>
          <div style={{ background: p.surface, borderRadius: 18, padding: '14px 18px', border: `1px dashed ${p.border}`, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <div style={{ fontSize: 12, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase' }}>Ended</div>
              <div className="kk-num" style={{ fontSize: 22, fontWeight: 500, color: p.inkMuted, marginTop: 2 }}>Pick a time</div>
            </div>
            <div style={{ fontSize: 13, color: p.inkMuted, fontWeight: 600 }}>Set</div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '0 4px', fontSize: 13, color: p.inkMuted }}>
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={p.inkMuted} strokeWidth="2" strokeLinecap="round"><circle cx="8" cy="8" r="6"/><path d="M8 5v3l2 1.5"/></svg>
            Duration appears once both times are set
          </div>
        </div>

        <div style={{ display: 'flex', gap: 10, marginTop: 22 }}>
          <BigButton palette={p} primary={false} style={{ flex: 1 }}>Cancel</BigButton>
          <div style={{
            flex: 1.4, padding: '14px 0', textAlign: 'center', borderRadius: 999,
            background: p.surface, color: p.inkMuted, fontWeight: 600, fontSize: 16,
            border: `1px solid ${p.border}`,
          }}>Save sleep</div>
        </div>
      </div>
    </PhoneShell>
  );
}

Object.assign(window, { Onboard1, Onboard2, Settings, LogPast, LogPastEmpty, LogPastDefault, EditSession });

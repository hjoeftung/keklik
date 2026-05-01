// Stats screens — Today, Week, Summary

function StatsHeader({ palette, baby = 'Sofia', tab = 'today', onTab }) {
  const p = palette;
  const tabs = [{ id: 'today', label: 'Today' }, { id: 'week', label: 'Week' }, { id: 'summary', label: 'Summary' }];
  return (
    <>
      <div style={{ paddingTop: 56, padding: '56px 24px 12px' }}>
        <div className="kk-display" style={{ fontSize: 28, fontWeight: 500, color: p.ink, letterSpacing: -0.5 }}>Sofia's sleep</div>
        <div style={{ fontSize: 13, color: p.inkSoft, marginTop: 2 }}>Tuesday · April 27</div>
      </div>
      <div style={{ padding: '8px 24px 0', display: 'flex', gap: 8 }}>
        {tabs.map(t => (
          <div key={t.id} style={{
            padding: '9px 18px', borderRadius: 999,
            background: tab === t.id ? p.ink : 'transparent',
            color: tab === t.id ? '#fff' : p.inkSoft,
            fontWeight: 600, fontSize: 14, border: tab === t.id ? 'none' : `1px solid ${p.border}`,
          }}>{t.label}</div>
        ))}
      </div>
    </>
  );
}

// ─── Today tab — vertical 20-hour timeline with rounded bars ───────────
function StatsToday({ palette }) {
  const p = palette;
  // 20-hour window from 06:00 to 02:00 next day
  // sample sessions
  const sessions = [
    { kind: 'nap', startH: 9.5, endH: 10.83, dur: '1h 20m' },
    { kind: 'nap', startH: 13.0, endH: 14.5, dur: '1h 30m' },
    { kind: 'night', startH: 19.78, endH: 25.5, dur: '5h 43m' }, // 19:47 → 01:30
  ];
  const winStart = 6, winEnd = 26; // 20h
  const TOTAL_H = 480; // px
  const hours = []; for (let h = winStart; h <= winEnd; h++) hours.push(h);
  const yFor = h => ((h - winStart) / (winEnd - winStart)) * TOTAL_H;

  return (
    <PhoneShell palette={p}>
      <StatsHeader palette={p} tab="today" />
      {/* Summary header */}
      <div style={{ padding: '18px 24px 8px', display: 'flex', gap: 10 }}>
        {[
          { label: 'Sleep', val: '5h 43m', color: p.night, soft: p.nightSoft },
          { label: 'Naps', val: '2h 50m', color: p.nap, soft: p.napSoft },
          { label: 'Active', val: '11h 27m', color: p.primary, soft: p.primarySoft },
        ].map(s => (
          <div key={s.label} style={{
            flex: 1, background: s.soft, borderRadius: 16, padding: '10px 12px',
          }}>
            <div style={{ fontSize: 11, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.4, textTransform: 'uppercase' }}>{s.label}</div>
            <div className="kk-num" style={{ fontSize: 17, fontWeight: 700, color: p.ink, marginTop: 2, letterSpacing: -0.3 }}>{s.val}</div>
          </div>
        ))}
      </div>
      {/* Timeline */}
      <div className="kk-scroll" style={{ position: 'absolute', top: 248, bottom: 80, left: 0, right: 0, padding: '12px 24px 24px' }}>
        <div style={{ position: 'relative', height: TOTAL_H, paddingLeft: 36 }}>
          {/* hour labels */}
          {hours.map(h => (
            <div key={h} style={{ position: 'absolute', left: 0, top: yFor(h) - 6, fontSize: 11, color: p.inkMuted, fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
              {String(h % 24).padStart(2, '0')}
            </div>
          ))}
          {/* hour grid lines */}
          {hours.map(h => (
            <div key={'l' + h} style={{ position: 'absolute', left: 28, right: 0, top: yFor(h), height: 1, background: p.border, opacity: 0.5 }}/>
          ))}
          {/* awake gaps between sessions — soft warm bands */}
          {sessions.map((s, i) => {
            const next = sessions[i + 1];
            if (!next) return null;
            const gapTop = yFor(s.endH);
            const gapH = yFor(next.startH) - gapTop;
            if (gapH < 14) return null;
            const gapDur = next.startH - s.endH;
            const hh = Math.floor(gapDur);
            const mm = Math.round((gapDur - hh) * 60);
            return (
              <div key={'g' + i} style={{
                position: 'absolute', left: 28, right: 0, top: gapTop, height: gapH,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: p.inkMuted, fontSize: 11, fontWeight: 600, letterSpacing: 0.3,
              }}>
                <span style={{
                  display: 'inline-flex', alignItems: 'center', gap: 6,
                  padding: '4px 10px', borderRadius: 999,
                  background: p.surfaceSoft, border: `1px dashed ${p.border}`,
                }}>
                  <svg width="11" height="11" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><circle cx="8" cy="8" r="3.5"/><path d="M8 1v2M8 13v2M1 8h2M13 8h2"/></svg>
                  awake · {hh > 0 ? `${hh}h ` : ''}{mm}m
                </span>
              </div>
            );
          })}
          {/* sessions */}
          {sessions.map((s, i) => {
            const top = yFor(s.startH);
            const height = yFor(s.endH) - top;
            const isNight = s.kind === 'night';
            return (
              <div key={i} style={{
                position: 'absolute', left: 28, right: 0, top, height,
                background: isNight ? p.night : p.nap,
                borderRadius: height < 30 ? 4 : (height < 80 ? 6 : 10), padding: height < 60 ? '6px 12px' : '8px 14px',
                display: 'flex', flexDirection: 'column', justifyContent: 'space-between',
                color: '#fff', boxShadow: `0 4px 12px ${isNight ? p.night : p.nap}40`,
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: 13, fontWeight: 700 }}>{isNight ? 'Night sleep' : 'Nap'}</span>
                  <span style={{ fontSize: 12, fontWeight: 600, opacity: 0.85 }} className="kk-num">{s.dur}</span>
                </div>
                {height > 60 && <div style={{ fontSize: 11, opacity: 0.8, fontWeight: 500 }} className="kk-num">
                  {String(Math.floor(s.startH)).padStart(2,'0')}:{String(Math.round((s.startH%1)*60)).padStart(2,'0')} – {String(Math.floor(s.endH)%24).padStart(2,'0')}:{String(Math.round((s.endH%1)*60)).padStart(2,'0')}
                </div>}
              </div>
            );
          })}
        </div>
      </div>
      <TabBar palette={p} active="stats" />
    </PhoneShell>
  );
}

// ─── Week tab — 7 columns × hour rows ──────────────────────────────────
function StatsWeek({ palette }) {
  const p = palette;
  const days = ['Wed 21','Thu 22','Fri 23','Sat 24','Sun 25','Mon 26','Tue 27'];
  // synthetic sessions per day [{startH, endH, kind}]
  const data = [
    [{s:9,e:10.5,k:'nap'},{s:13,e:14.5,k:'nap'},{s:19.5,e:25.7,k:'night'}],
    [{s:9.7,e:11,k:'nap'},{s:13.2,e:14.4,k:'nap'},{s:20,e:25.2,k:'night'}],
    [{s:8.8,e:10,k:'nap'},{s:13,e:14.7,k:'nap'},{s:19.8,e:25.9,k:'night'}],
    [{s:10,e:11,k:'nap'},{s:14,e:15.3,k:'nap'},{s:20.2,e:25.5,k:'night'}],
    [{s:9.5,e:11.2,k:'nap'},{s:13.5,e:14.8,k:'nap'},{s:19.7,e:25.6,k:'night'}],
    [{s:9,e:10.3,k:'nap'},{s:13,e:14.5,k:'nap'},{s:19.6,e:25.4,k:'night'}],
    [{s:9.5,e:10.83,k:'nap'},{s:13,e:14.5,k:'nap'},{s:19.78,e:25.5,k:'night'}],
  ];
  const winStart = 6, winEnd = 26;
  const COL_H = 460;
  const yFor = h => ((h - winStart) / (winEnd - winStart)) * COL_H;
  const hourTicks = [6, 12, 18, 24];

  return (
    <PhoneShell palette={p}>
      <StatsHeader palette={p} tab="week" />
      <div style={{ padding: '18px 16px 0' }}>
        {/* Legend */}
        <div style={{ display: 'flex', gap: 12, padding: '0 8px 12px' }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, fontWeight: 600, color: p.inkSoft }}>
            <span style={{ width: 10, height: 10, borderRadius: 3, background: p.night }}/>Night
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, fontWeight: 600, color: p.inkSoft }}>
            <span style={{ width: 10, height: 10, borderRadius: 3, background: p.nap }}/>Nap
          </span>
        </div>
        <div style={{ display: 'flex', gap: 4, paddingLeft: 24 }}>
          {days.map(d => (
            <div key={d} style={{ flex: 1, fontSize: 10, fontWeight: 600, color: p.inkSoft, textAlign: 'center' }}>
              {d.split(' ')[0]}<br/><span style={{ color: p.ink, fontSize: 12, fontWeight: 700 }}>{d.split(' ')[1]}</span>
            </div>
          ))}
        </div>
        <div style={{ position: 'relative', marginTop: 8, height: COL_H, paddingLeft: 24 }}>
          {/* hour ticks */}
          {hourTicks.map(h => (
            <React.Fragment key={h}>
              <div style={{ position: 'absolute', left: 0, top: yFor(h) - 5, fontSize: 10, fontWeight: 600, color: p.inkMuted }} className="kk-num">{String(h%24).padStart(2,'0')}</div>
              <div style={{ position: 'absolute', left: 22, right: 0, top: yFor(h), height: 1, background: p.border, opacity: 0.5 }}/>
            </React.Fragment>
          ))}
          {/* day columns */}
          <div style={{ display: 'flex', gap: 4, height: COL_H }}>
            {data.map((day, di) => (
              <div key={di} style={{ flex: 1, position: 'relative' }}>
                {day.map((s, si) => {
                  const dur = s.e - s.s;
                  return (
                    <div key={si} style={{
                      position: 'absolute', left: 1, right: 1,
                      top: yFor(s.s) + 1, height: yFor(s.e) - yFor(s.s) - 2,
                      background: s.k === 'night' ? p.night : p.nap,
                      borderRadius: dur < 0.6 ? 1.5 : (dur < 1.2 ? 2.5 : 4),
                    }}/>
                  );
                })}
              </div>
            ))}
          </div>
        </div>
      </div>
      <TabBar palette={p} active="stats" />
    </PhoneShell>
  );
}

// ─── Summary tab — period selector + 3 averages ────────────────────────
function StatsSummary({ palette }) {
  const p = palette;
  const periods = ['7d', '14d', '30d', '90d'];
  return (
    <PhoneShell palette={p}>
      <StatsHeader palette={p} tab="summary" />
      <div style={{ padding: '18px 24px 0' }}>
        <div style={{ display: 'flex', gap: 6, background: p.surfaceSoft, padding: 4, borderRadius: 999 }}>
          {periods.map((pp, i) => (
            <div key={pp} style={{
              flex: 1, padding: '8px 0', textAlign: 'center', borderRadius: 999,
              background: i === 0 ? p.surface : 'transparent',
              color: i === 0 ? p.ink : p.inkSoft,
              fontWeight: 600, fontSize: 14,
              boxShadow: i === 0 ? '0 2px 6px rgba(0,0,0,0.06)' : 'none',
            }}>{pp}</div>
          ))}
        </div>
        <div style={{ marginTop: 14, fontSize: 13, color: p.inkSoft, textAlign: 'center' }}>
          21 April – 27 April 2026
        </div>
        <div style={{ marginTop: 28, display: 'flex', flexDirection: 'column', gap: 14 }}>
          {[
            { label: 'Avg sleep', val: '5h 38m', sub: 'per night', color: p.night, soft: p.nightSoft, icon: 'moon' },
            { label: 'Avg nap', val: '2h 47m', sub: 'per day', color: p.nap, soft: p.napSoft, icon: 'sun' },
            { label: 'Avg active', val: '11h 35m', sub: 'per day', color: p.primary, soft: p.primarySoft, icon: 'spark' },
          ].map(row => (
            <div key={row.label} style={{
              background: row.soft, borderRadius: 24, padding: 20,
              display: 'flex', alignItems: 'center', gap: 16,
            }}>
              <div style={{
                width: 56, height: 56, borderRadius: 18, background: row.color,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                {row.icon === 'moon' && <Moon size={32} palette={p} mode="crescent" />}
                {row.icon === 'sun' && <svg width="30" height="30" viewBox="0 0 30 30" fill="#fff"><circle cx="15" cy="15" r="6"/><g stroke="#fff" strokeWidth="2.2" strokeLinecap="round"><path d="M15 4v3M15 23v3M4 15h3M23 15h3M7 7l2 2M21 21l2 2M7 23l2-2M21 9l2-2"/></g></svg>}
                {row.icon === 'spark' && <svg width="30" height="30" viewBox="0 0 30 30" fill="#fff"><path d="M15 5l2.5 7.5L25 15l-7.5 2.5L15 25l-2.5-7.5L5 15l7.5-2.5z"/></svg>}
              </div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 13, fontWeight: 700, color: p.inkSoft, letterSpacing: 0.3, textTransform: 'uppercase' }}>{row.label}</div>
                <div className="kk-num kk-display" style={{ fontSize: 32, fontWeight: 500, color: p.ink, lineHeight: 1.05, marginTop: 4, letterSpacing: -0.6 }}>{row.val}</div>
                <div style={{ fontSize: 12, color: p.inkSoft, marginTop: 2 }}>{row.sub}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
      <TabBar palette={p} active="stats" />
    </PhoneShell>
  );
}

Object.assign(window, { StatsToday, StatsWeek, StatsSummary });

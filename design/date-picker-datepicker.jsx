
// ─────────────────────────────────────────────────────────────────────────────
// Keklik · Date Picker · two variations
//
// Variation A — Drum Roll
//   Three-column slot-machine scrollers: [Day] [Hour] [Minute] (+ optional AM/PM)
//   Lives inside the existing LogPastSleepSheet panel. No modal, no calendar.
//   Warm, haptic-feeling columns with the Keklik palette.
//
// Variation B — Calendar + Time
//   Compact month grid for date selection, then a two-column drum for hour/min.
//   Slides open like a tray inside the sheet.
// ─────────────────────────────────────────────────────────────────────────────

// ─── Shared picker helpers ────────────────────────────────────────────────────

const ITEM_H = 44; // row height in drum columns

/**
 * DrumColumn — vertically scrolling column of items.
 * Snaps to rows with momentum, highlights the centre item.
 */
function DrumColumn({ items, value, onChange, palette, width = 80, label }) {
  const p = palette;
  const ref = React.useRef(null);
  const dragging = React.useRef(false);
  const startY = React.useRef(0);
  const startScrollTop = React.useRef(0);
  const scrollTimeout = React.useRef(null);

  const visibleRows = 5; // must be odd
  const viewH = visibleRows * ITEM_H;
  const padItems = Math.floor(visibleRows / 2);

  // Padded items so selection can scroll to top/bottom ends
  const padded = [
    ...Array(padItems).fill(null),
    ...items,
    ...Array(padItems).fill(null),
  ];

  const selectedIdx = items.indexOf(value);

  // Scroll to selectedIdx on mount + whenever value changes from outside
  const ignoreScroll = React.useRef(false);
  React.useEffect(() => {
    if (!ref.current) return;
    ignoreScroll.current = true;
    ref.current.scrollTop = selectedIdx * ITEM_H;
    setTimeout(() => { ignoreScroll.current = false; }, 50);
  }, [value, selectedIdx]);

  function snapToNearest() {
    if (!ref.current) return;
    const raw = ref.current.scrollTop;
    const idx = Math.round(raw / ITEM_H);
    const clamped = Math.max(0, Math.min(idx, items.length - 1));
    ref.current.scrollTop = clamped * ITEM_H;
    if (items[clamped] !== value) onChange(items[clamped]);
  }

  function handleScroll() {
    if (ignoreScroll.current) return;
    clearTimeout(scrollTimeout.current);
    scrollTimeout.current = setTimeout(snapToNearest, 120);
  }

  // Touch drag for mobile
  function onTouchStart(e) {
    startY.current = e.touches[0].clientY;
    startScrollTop.current = ref.current.scrollTop;
  }
  function onTouchMove(e) {
    const dy = startY.current - e.touches[0].clientY;
    ref.current.scrollTop = startScrollTop.current + dy;
  }
  function onTouchEnd() {
    snapToNearest();
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
      {label && (
        <div style={{
          fontSize: 10, fontWeight: 700, letterSpacing: 0.6, textTransform: 'uppercase',
          color: p.inkMuted, marginBottom: 2,
        }}>{label}</div>
      )}
      <div style={{ position: 'relative', width, height: viewH, overflow: 'hidden' }}>
        {/* top fade */}
        <div style={{
          position: 'absolute', top: 0, left: 0, right: 0, height: ITEM_H * 2,
          background: `linear-gradient(to bottom, ${p.surface}, transparent)`,
          pointerEvents: 'none', zIndex: 2,
        }} />
        {/* bottom fade */}
        <div style={{
          position: 'absolute', bottom: 0, left: 0, right: 0, height: ITEM_H * 2,
          background: `linear-gradient(to top, ${p.surface}, transparent)`,
          pointerEvents: 'none', zIndex: 2,
        }} />
        {/* selection highlight band */}
        <div style={{
          position: 'absolute', top: padItems * ITEM_H, left: 6, right: 6,
          height: ITEM_H, borderRadius: 12,
          background: p.primarySoft,
          border: `1.5px solid ${p.primary}40`,
          pointerEvents: 'none', zIndex: 1,
        }} />
        {/* scrollable list */}
        <div
          ref={ref}
          onScroll={handleScroll}
          onTouchStart={onTouchStart}
          onTouchMove={onTouchMove}
          onTouchEnd={onTouchEnd}
          style={{
            height: '100%', overflowY: 'scroll', scrollbarWidth: 'none',
            WebkitOverflowScrolling: 'touch',
          }}
        >
          <style>{`div::-webkit-scrollbar{display:none}`}</style>
          {padded.map((item, i) => {
            const isSelected = item !== null && item === value;
            return (
              <div
                key={i}
                onClick={() => item !== null && (ref.current.scrollTop = items.indexOf(item) * ITEM_H, onChange(item))}
                style={{
                  height: ITEM_H, display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: isSelected ? 22 : 17,
                  fontWeight: isSelected ? 700 : 500,
                  color: isSelected ? p.primaryDeep : p.inkSoft,
                  fontFamily: '"Quicksand", system-ui',
                  fontVariantNumeric: 'tabular-nums',
                  cursor: item !== null ? 'pointer' : 'default',
                  transition: 'font-size 0.1s, color 0.1s',
                  userSelect: 'none',
                  position: 'relative', zIndex: 3,
                  letterSpacing: isSelected ? '-0.5px' : '0',
                }}
              >
                {item !== null ? item : ''}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

// Build day items for the last N days: "Today", "Yesterday", "Mon 28", etc.
function buildDayItems(maxDaysBack = 30) {
  const items = [];
  const today = new Date();
  for (let i = 0; i <= maxDaysBack; i++) {
    const d = new Date(today);
    d.setDate(today.getDate() - i);
    items.push({ date: d, label: i === 0 ? 'Today' : i === 1 ? 'Yesterday' : formatDayLabel(d) });
  }
  return items;
}

function formatDayLabel(d) {
  return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
}

function pad2(n) { return String(n).padStart(2, '0'); }

function buildHours(is24) {
  if (is24) return Array.from({ length: 24 }, (_, i) => pad2(i));
  return Array.from({ length: 12 }, (_, i) => pad2(i === 0 ? 12 : i));
}

function buildMinutes(step = 1) {
  return Array.from({ length: 60 / step }, (_, i) => pad2(i * step));
}

// ─── VARIATION A: Pure drum roll ──────────────────────────────────────────────

function PickerDrumRoll({ palette, initialDate, is24h, onChange }) {
  const p = palette;

  const dayItems = React.useMemo(() => buildDayItems(30), []);
  const hourItems = React.useMemo(() => buildHours(is24h), [is24h]);
  const minItems = React.useMemo(() => buildMinutes(1), []);
  const ampmItems = ['AM', 'PM'];

  const init = initialDate || new Date();
  const initDayLabel = (() => {
    const today = new Date(); const yest = new Date(); yest.setDate(today.getDate()-1);
    if (init.toDateString() === today.toDateString()) return 'Today';
    if (init.toDateString() === yest.toDateString()) return 'Yesterday';
    return formatDayLabel(init);
  })();

  const [day, setDay] = React.useState(initDayLabel);
  const [hour, setHour] = React.useState(() => {
    const h = init.getHours();
    return is24h ? pad2(h) : pad2(h === 0 ? 12 : h > 12 ? h - 12 : h);
  });
  const [minute, setMinute] = React.useState(pad2(init.getMinutes()));
  const [ampm, setAmpm] = React.useState(init.getHours() >= 12 ? 'PM' : 'AM');

  // Notify parent of changes
  React.useEffect(() => {
    const dayObj = dayItems.find(d => d.label === day);
    if (!dayObj || !hour || !minute) return;
    const result = new Date(dayObj.date);
    let h = parseInt(hour, 10);
    if (!is24h) {
      if (ampm === 'PM' && h !== 12) h += 12;
      if (ampm === 'AM' && h === 12) h = 0;
    }
    result.setHours(h, parseInt(minute, 10), 0, 0);
    onChange && onChange(result);
  }, [day, hour, minute, ampm, is24h]);

  const dayLabels = dayItems.map(d => d.label);

  return (
    <div style={{
      background: p.surface,
      borderRadius: 22,
      padding: '16px 12px 20px',
      border: `1px solid ${p.border}`,
      display: 'flex',
      flexDirection: 'column',
      gap: 0,
    }}>
      {/* columns */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: is24h ? 4 : 2,
      }}>
        {/* Day — wider */}
        <DrumColumn items={dayLabels} value={day} onChange={setDay} palette={p} width={130} label="Day" />

        {/* divider */}
        <div style={{ fontSize: 22, fontWeight: 600, color: p.inkMuted, margin: '20px 2px 0' }}>·</div>

        {/* Hour */}
        <DrumColumn items={hourItems} value={hour} onChange={setHour} palette={p} width={56} label="Hour" />

        {/* colon */}
        <div style={{ fontSize: 22, fontWeight: 700, color: p.primaryDeep, margin: '20px 0px 0', letterSpacing: 0 }}>:</div>

        {/* Minute */}
        <DrumColumn items={minItems} value={minute} onChange={setMinute} palette={p} width={56} label="Min" />

        {/* AM/PM column if 12h */}
        {!is24h && (
          <>
            <div style={{ width: 8 }} />
            <DrumColumn items={ampmItems} value={ampm} onChange={setAmpm} palette={p} width={52} label="" />
          </>
        )}
      </div>
    </div>
  );
}

// ─── VARIATION B: Mini calendar + time scroller ───────────────────────────────

function MiniCalendar({ palette, value, onChange }) {
  const p = palette;
  const today = new Date();

  const [viewYear, setViewYear] = React.useState(value ? value.getFullYear() : today.getFullYear());
  const [viewMonth, setViewMonth] = React.useState(value ? value.getMonth() : today.getMonth());

  const firstDay = new Date(viewYear, viewMonth, 1).getDay(); // 0=Sun
  const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate();
  const prevMonthDays = new Date(viewYear, viewMonth, 0).getDate();

  const monthLabel = new Date(viewYear, viewMonth).toLocaleDateString('en-US', { month: 'long', year: 'numeric' });

  // Max date = today
  function isDisabled(d) {
    const dt = new Date(viewYear, viewMonth, d);
    return dt > today;
  }

  function isSelected(d) {
    if (!value) return false;
    return value.getFullYear() === viewYear && value.getMonth() === viewMonth && value.getDate() === d;
  }

  function isToday(d) {
    return today.getFullYear() === viewYear && today.getMonth() === viewMonth && today.getDate() === d;
  }

  function prevMonth() {
    if (viewMonth === 0) { setViewYear(y => y - 1); setViewMonth(11); }
    else setViewMonth(m => m - 1);
  }
  function nextMonth() {
    const next = new Date(viewYear, viewMonth + 1, 1);
    if (next > today) return; // don't go into future
    if (viewMonth === 11) { setViewYear(y => y + 1); setViewMonth(0); }
    else setViewMonth(m => m + 1);
  }

  // Build 6-row grid
  const cells = [];
  // leading blanks from previous month
  for (let i = 0; i < firstDay; i++) {
    cells.push({ day: prevMonthDays - firstDay + 1 + i, type: 'prev' });
  }
  for (let d = 1; d <= daysInMonth; d++) {
    cells.push({ day: d, type: 'cur' });
  }
  while (cells.length % 7 !== 0) {
    cells.push({ day: cells.length - daysInMonth - firstDay + 1, type: 'next' });
  }

  const weeks = [];
  for (let i = 0; i < cells.length; i += 7) weeks.push(cells.slice(i, i + 7));

  return (
    <div style={{ padding: '0 4px' }}>
      {/* Month nav */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
        <button onClick={prevMonth} style={{
          background: 'none', border: 'none', cursor: 'pointer', padding: '4px 8px',
          color: p.primaryDeep, fontSize: 18, lineHeight: 1,
        }}>‹</button>
        <span style={{ fontSize: 14, fontWeight: 700, color: p.ink, letterSpacing: -0.2 }}>{monthLabel}</span>
        <button onClick={nextMonth} style={{
          background: 'none', border: 'none', cursor: 'pointer', padding: '4px 8px',
          color: new Date(viewYear, viewMonth + 1, 1) > today ? p.inkMuted : p.primaryDeep,
          fontSize: 18, lineHeight: 1,
        }}>›</button>
      </div>

      {/* Weekday headers */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', marginBottom: 4 }}>
        {['Su','Mo','Tu','We','Th','Fr','Sa'].map(d => (
          <div key={d} style={{
            textAlign: 'center', fontSize: 11, fontWeight: 700, color: p.inkMuted,
            letterSpacing: 0.4, paddingBottom: 4,
          }}>{d}</div>
        ))}
      </div>

      {/* Day grid */}
      {weeks.map((week, wi) => (
        <div key={wi} style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: '2px 0' }}>
          {week.map((cell, ci) => {
            const isCur = cell.type === 'cur';
            const sel = isCur && isSelected(cell.day);
            const tod = isCur && isToday(cell.day);
            const dis = !isCur || isDisabled(cell.day);
            return (
              <button
                key={ci}
                disabled={dis}
                onClick={() => {
                  if (dis) return;
                  const newDate = value ? new Date(value) : new Date();
                  newDate.setFullYear(viewYear, viewMonth, cell.day);
                  onChange(newDate);
                }}
                style={{
                  width: '100%', aspectRatio: '1', borderRadius: 10, border: 'none',
                  background: sel ? p.primary : 'transparent',
                  color: dis ? p.inkMuted : sel ? '#fff' : tod ? p.primaryDeep : p.ink,
                  fontSize: 13, fontWeight: sel ? 700 : tod ? 700 : 500,
                  cursor: dis ? 'default' : 'pointer',
                  opacity: !isCur ? 0.3 : 1,
                  position: 'relative',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  outline: tod && !sel ? `2px solid ${p.primary}` : 'none',
                  outlineOffset: '-2px',
                  transition: 'background 0.1s',
                  fontFamily: '"Quicksand", system-ui',
                }}
              >{cell.day}</button>
            );
          })}
        </div>
      ))}
    </div>
  );
}

function PickerCalendarTime({ palette, initialDate, is24h, onChange }) {
  const p = palette;

  const init = initialDate || new Date();
  const [date, setDate] = React.useState(init);

  const hourItems = React.useMemo(() => buildHours(is24h), [is24h]);
  const minItems = React.useMemo(() => buildMinutes(1), []);
  const ampmItems = ['AM', 'PM'];

  const [hour, setHour] = React.useState(() => {
    const h = init.getHours();
    return is24h ? pad2(h) : pad2(h === 0 ? 12 : h > 12 ? h - 12 : h);
  });
  const [minute, setMinute] = React.useState(pad2(init.getMinutes()));
  const [ampm, setAmpm] = React.useState(init.getHours() >= 12 ? 'PM' : 'AM');

  React.useEffect(() => {
    if (!date) return;
    const result = new Date(date);
    let h = parseInt(hour, 10);
    if (!is24h) {
      if (ampm === 'PM' && h !== 12) h += 12;
      if (ampm === 'AM' && h === 12) h = 0;
    }
    result.setHours(h, parseInt(minute, 10), 0, 0);
    onChange && onChange(result);
  }, [date, hour, minute, ampm, is24h]);

  const dateLabel = (() => {
    const today = new Date(); const yest = new Date(); yest.setDate(today.getDate()-1);
    if (date.toDateString() === today.toDateString()) return 'Today';
    if (date.toDateString() === yest.toDateString()) return 'Yesterday';
    return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
  })();

  return (
    <div style={{
      background: p.surface,
      borderRadius: 22,
      border: `1px solid ${p.border}`,
      overflow: 'hidden',
    }}>
      {/* Calendar section */}
      <div style={{ padding: '16px 16px 12px' }}>
        <MiniCalendar palette={p} value={date} onChange={setDate} />
      </div>

      {/* Divider */}
      <div style={{ height: 1, background: p.border, margin: '0 16px' }} />

      {/* Time drums */}
      <div style={{ padding: '12px 12px 16px' }}>
        <div style={{
          fontSize: 10, fontWeight: 700, letterSpacing: 0.6, textTransform: 'uppercase',
          color: p.inkMuted, marginBottom: 8, textAlign: 'center',
        }}>Time</div>
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          gap: is24h ? 4 : 2,
        }}>
          <DrumColumn items={hourItems} value={hour} onChange={setHour} palette={p} width={64} />
          <div style={{ fontSize: 22, fontWeight: 700, color: p.primaryDeep, margin: '0 2px', paddingBottom: 0 }}>:</div>
          <DrumColumn items={minItems} value={minute} onChange={setMinute} palette={p} width={64} />
          {!is24h && (
            <>
              <div style={{ width: 8 }} />
              <DrumColumn items={ampmItems} value={ampm} onChange={setAmpm} palette={p} width={56} />
            </>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Sheet wrapper — mimics LogPastSleepSheet ─────────────────────────────────

function LogSleepSheet({ palette, pickerVariant = 'drum', is24h = true, sleepType = 'nap', label }) {
  const p = palette;

  const defaultEnd = new Date();
  defaultEnd.setSeconds(0, 0);
  const defaultStart = new Date(defaultEnd.getTime() - 60 * 60 * 1000);

  const [activePicker, setActivePicker] = React.useState(null); // 'start' | 'end' | null
  const [startDate, setStartDate] = React.useState(defaultStart);
  const [endDate, setEndDate] = React.useState(defaultEnd);
  const [sleepTypeState, setSleepType] = React.useState(sleepType);

  const durSecs = (endDate - startDate) / 1000;
  const isError = durSecs <= 0;

  function formatDisplayTime(d) {
    const today = new Date();
    const yest = new Date(); yest.setDate(today.getDate()-1);
    const timeStr = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: !is24h });
    if (d.toDateString() === today.toDateString()) return `Today, ${timeStr}`;
    if (d.toDateString() === yest.toDateString()) return `Yesterday, ${timeStr}`;
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + `, ${timeStr}`;
  }

  function formatDur(s) {
    const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60);
    return h > 0 ? `${h}h ${m}m` : `${m}m`;
  }

  const PickerComponent = pickerVariant === 'drum' ? PickerDrumRoll : PickerCalendarTime;

  return (
    <div className="kk" style={{
      width: 390, height: 780,
      position: 'relative', background: p.bg,
      borderRadius: 44, overflow: 'hidden',
      boxShadow: '0 20px 50px rgba(60,40,20,0.18), 0 0 0 1px rgba(0,0,0,0.06)',
      fontFamily: '"Quicksand", system-ui',
    }}>
      {/* status bar */}
      <div style={{
        position: 'absolute', top: 0, left: 0, right: 0, zIndex: 30,
        height: 44, padding: '14px 28px 0', display: 'flex',
        justifyContent: 'space-between', alignItems: 'center',
        color: p.ink, fontFamily: '"Quicksand", system-ui', fontSize: 15, fontWeight: 600,
      }}>
        <span>9:41</span>
        <div style={{ display: 'flex', gap: 6, alignItems: 'center', opacity: 0.85 }}>
          <svg width="16" height="11" viewBox="0 0 16 11" fill="currentColor"><rect x="0" y="7" width="3" height="4" rx="0.5"/><rect x="4.5" y="5" width="3" height="6" rx="0.5"/><rect x="9" y="2.5" width="3" height="8.5" rx="0.5"/><rect x="13.5" y="0" width="3" height="11" rx="0.5"/></svg>
          <svg width="22" height="11" viewBox="0 0 22 11" fill="none" stroke="currentColor" strokeWidth="1"><rect x="0.5" y="0.5" width="18" height="10" rx="2.5"/><rect x="2" y="2" width="15" height="7" rx="1.2" fill="currentColor"/><rect x="20" y="3.5" width="1.5" height="4" rx="0.5" fill="currentColor"/></svg>
        </div>
      </div>
      {/* dynamic island */}
      <div style={{ position: 'absolute', top: 11, left: '50%', transform: 'translateX(-50%)', width: 110, height: 32, borderRadius: 20, background: '#000', zIndex: 40 }} />

      {/* main sheet slide-up */}
      <div style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        background: p.bg,
        borderRadius: '28px 28px 0 0',
        padding: '12px 20px 28px',
        maxHeight: 700,
        overflowY: 'auto',
        scrollbarWidth: 'none',
      }}>
        {/* handle */}
        <div style={{ width: 44, height: 5, borderRadius: 3, background: p.border, margin: '0 auto 14px' }} />

        <div style={{ fontSize: 20, fontWeight: 600, color: p.ink, fontFamily: '"Fraunces", Georgia, serif', letterSpacing: -0.3 }}>Log past sleep</div>
        <div style={{ fontSize: 13, color: p.inkSoft, marginTop: 3 }}>We've suggested times — adjust if needed</div>

        {/* Sleep type toggle */}
        <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
          {[{ id: 'night', label: 'Night sleep' }, { id: 'nap', label: 'Nap' }].map(({ id, label: lbl }) => (
            <button key={id} onClick={() => setSleepType(id)} style={{
              flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
              padding: '11px 14px', borderRadius: 16, border: 'none', cursor: 'pointer',
              fontFamily: 'inherit', fontSize: 14, fontWeight: 600,
              background: sleepTypeState === id ? (id === 'night' ? p.night : p.nap) : p.surface,
              color: sleepTypeState === id ? '#fff' : p.inkSoft,
              border: sleepTypeState === id ? 'none' : `1px solid ${p.border}`,
              transition: 'all 0.15s',
            }}>{lbl}</button>
          ))}
        </div>

        {/* Time rows */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 14 }}>
          {['start', 'end'].map(which => {
            const isActive = activePicker === which;
            const dateVal = which === 'start' ? startDate : endDate;
            return (
              <React.Fragment key={which}>
                <button
                  onClick={() => setActivePicker(isActive ? null : which)}
                  style={{
                    display: 'block', width: '100%', textAlign: 'left', cursor: 'pointer',
                    background: isActive ? p.primarySoft : p.surface,
                    border: isActive ? `1.5px solid ${p.primary}` : `1px solid ${p.border}`,
                    borderRadius: 18, padding: '13px 16px',
                    fontFamily: 'inherit',
                    outline: 'none',
                    transition: 'all 0.15s',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <div>
                      <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: 0.5, textTransform: 'uppercase', color: isActive ? p.primaryDeep : p.inkMuted }}>
                        {which === 'start' ? 'STARTED' : 'ENDED'}{which === 'end' && !isActive ? ' · NOW' : ''}
                      </div>
                      <div style={{ fontSize: 20, fontWeight: 600, color: p.ink, marginTop: 2, fontVariantNumeric: 'tabular-nums' }}>
                        {formatDisplayTime(dateVal)}
                      </div>
                    </div>
                    <span style={{ fontSize: 13, fontWeight: 600, color: isActive ? p.primaryDeep : p.primary }}>
                      {isActive ? 'Done ✓' : 'Change'}
                    </span>
                  </div>
                </button>

                {/* Inline picker — expands below the row */}
                {isActive && (
                  <div style={{ marginTop: -4 }}>
                    <PickerComponent
                      palette={p}
                      initialDate={dateVal}
                      is24h={is24h}
                      onChange={d => which === 'start' ? setStartDate(d) : setEndDate(d)}
                    />
                  </div>
                )}
              </React.Fragment>
            );
          })}

          {/* Duration pill */}
          <div style={{
            display: 'flex', alignItems: 'center', gap: 8,
            padding: '2px 4px', fontSize: 13, color: p.inkSoft, minHeight: 22,
          }}>
            {isError ? (
              <>
                <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="#D4806E" strokeWidth="2" strokeLinecap="round"><circle cx="8" cy="8" r="6.5"/><path d="M8 5v3M8 11h.01"/></svg>
                <span style={{ color: '#D4806E' }}>End must be after start</span>
              </>
            ) : (
              <>
                <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="#86B6A6" strokeWidth="2" strokeLinecap="round"><path d="M3 8l3 3 7-7"/></svg>
                <span>Duration <strong style={{ color: p.ink, fontVariantNumeric: 'tabular-nums' }}>{formatDur(durSecs)}</strong></span>
              </>
            )}
          </div>
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', gap: 10, marginTop: 18 }}>
          <button style={{
            flex: 1, padding: '14px', borderRadius: 999, background: 'none',
            border: `1.5px solid ${p.border}`, fontFamily: 'inherit', fontSize: 15,
            fontWeight: 500, color: p.inkSoft, cursor: 'pointer',
          }}>Cancel</button>
          <button style={{
            flex: 1.4, padding: '14px', borderRadius: 999, border: 'none',
            background: isError ? p.border : p.primary,
            fontFamily: 'inherit', fontSize: 15, fontWeight: 600,
            color: '#fff', cursor: 'pointer',
            boxShadow: isError ? 'none' : `0 4px 12px ${p.primary}50`,
            transition: 'background 0.2s',
          }}>Save sleep</button>
        </div>

        {/* label badge */}
        {label && (
          <div style={{
            position: 'absolute', top: 14, right: 18,
            background: p.accentSoft, color: p.accent, fontSize: 11, fontWeight: 700,
            letterSpacing: 0.4, borderRadius: 8, padding: '3px 8px',
          }}>{label}</div>
        )}
      </div>

      {/* home indicator */}
      <div style={{ position: 'absolute', bottom: 8, left: 0, right: 0, zIndex: 50, display: 'flex', justifyContent: 'center', pointerEvents: 'none' }}>
        <div style={{ width: 134, height: 5, borderRadius: 3, background: 'rgba(46,42,51,0.25)' }} />
      </div>
    </div>
  );
}

// Export everything
Object.assign(window, { LogSleepSheet, PickerDrumRoll, PickerCalendarTime, DrumColumn, MiniCalendar });

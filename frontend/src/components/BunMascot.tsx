export default function BunMascot({ size = 104, sleepy = false }: { size?: number; sleepy?: boolean }) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 120 120">
      <g fill="#FBE4D1" stroke="#C97B4D" strokeWidth="1.4" strokeOpacity="0.3">
        <path d="M 58 36 C 56 28, 50 24, 48 20 Q 54 26, 60 28 Q 60 32, 58 36 Z" />
        <path d="M 64 35 C 70 28, 78 28, 80 22 Q 76 30, 72 32 Q 70 36, 64 35 Z" />
        <path d="M 50 38 C 46 34, 40 36, 38 32 Q 44 36, 48 36 Q 50 37, 50 38 Z" />
      </g>
      <path
        d="M 60 36 C 38 36, 24 56, 24 78 C 24 100, 40 116, 60 116 C 80 116, 96 100, 96 78 C 96 56, 82 36, 60 36 Z"
        fill="#FBE4D1"
        stroke="#E59B6A"
        strokeWidth="2"
        strokeOpacity="0.22"
      />
      <path
        d="M 28 74 Q 18 70, 16 64 Q 20 70, 24 72 Q 18 74, 20 78 Q 24 76, 28 76 Z"
        fill="#FBE4D1"
        stroke="#E59B6A"
        strokeWidth="1.4"
        strokeOpacity="0.25"
      />
      <path
        d="M 92 74 Q 102 70, 104 64 Q 100 70, 96 72 Q 102 74, 100 78 Q 96 76, 92 76 Z"
        fill="#FBE4D1"
        stroke="#E59B6A"
        strokeWidth="1.4"
        strokeOpacity="0.25"
      />
      <ellipse cx="40" cy="76" rx="5.5" ry="3.6" fill="#E59B6A" opacity="0.35" />
      <ellipse cx="78" cy="76" rx="5.5" ry="3.6" fill="#E59B6A" opacity="0.35" />
      {sleepy ? (
        <g fill="none" stroke="#2E2A33" strokeWidth="2" strokeLinecap="round">
          <path d="M 46 70 Q 49 67, 52 70" />
          <path d="M 68 70 Q 71 67, 74 70" />
        </g>
      ) : (
        <g>
          <ellipse cx="49" cy="69" rx="3.2" ry="3.6" fill="#2E2A33" />
          <ellipse cx="71" cy="69" rx="3.2" ry="3.6" fill="#2E2A33" />
          <circle cx="50" cy="67.5" r="1" fill="#fff" />
          <circle cx="72" cy="67.5" r="1" fill="#fff" />
        </g>
      )}
      <path
        d="M 56 78 L 60 84 L 64 78 Q 60 76, 56 78 Z"
        fill="#E8B86E"
        stroke="#C97B4D"
        strokeWidth="1.2"
        strokeOpacity="0.4"
        strokeLinejoin="round"
      />
      {sleepy && (
        <g opacity="0.7">
          <text x="96" y="46" fontFamily="Fraunces, Georgia, serif" fontSize="13" fontWeight="700" fill="#A38FC4">z</text>
          <text x="104" y="38" fontFamily="Fraunces, Georgia, serif" fontSize="10" fontWeight="700" fill="#A38FC4">z</text>
        </g>
      )}
    </svg>
  )
}

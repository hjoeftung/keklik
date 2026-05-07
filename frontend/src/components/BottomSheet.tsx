import { useState, useEffect } from 'react'
import { createPortal } from 'react-dom'
import styles from './BottomSheet.module.css'

interface BottomSheetProps {
  title: string
  onClose: () => void
  children: React.ReactNode
}

export function BottomSheet({ title, onClose, children }: BottomSheetProps) {
  const [closing, setClosing] = useState(false)

  useEffect(() => {
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = prev
    }
  }, [])

  function handleClose() {
    setClosing(true)
    setTimeout(onClose, 260)
  }

  return createPortal(
    <div
      className={closing ? `${styles.overlay} ${styles.overlayClosing}` : styles.overlay}
      onClick={handleClose}
    >
      <div
        className={closing ? `${styles.sheet} ${styles.sheetClosing}` : styles.sheet}
        onClick={(e) => e.stopPropagation()}
      >
        <div className={styles.handle} />
        <div className={styles.header}>
          <span className={styles.title}>{title}</span>
          <button type="button" className={styles.done} onClick={handleClose}>
            Done
          </button>
        </div>
        <div className={styles.body}>{children}</div>
      </div>
    </div>,
    document.body,
  )
}

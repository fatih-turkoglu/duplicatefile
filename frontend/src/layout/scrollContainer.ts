import { createContext, useContext } from 'react'

// Ana içerik alanının kaydırılan öğesi. Sanal liste kendi kaydırma kutusunu
// açmak yerine sayfayla birlikte kaysın diye bu öğeyi kullanır.
export const ScrollContainerContext = createContext<React.RefObject<HTMLElement | null> | null>(null)

export function useScrollContainer(): React.RefObject<HTMLElement | null> {
  const ref = useContext(ScrollContainerContext)
  if (!ref) throw new Error('useScrollContainer, AppLayout içinde kullanılmalı')
  return ref
}

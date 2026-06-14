import type { LinkProps } from 'next/link'
import type { ReactNode } from 'react'

export default function MockLink({
  href,
  children,
  ...props
}: LinkProps & { children?: ReactNode }) {
  return (
    <a href={typeof href === 'string' ? href : String(href)} {...props}>
      {children}
    </a>
  )
}

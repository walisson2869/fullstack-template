import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import Home from '../app/page'

describe('Home page', () => {
  it('renders the getting-started heading', () => {
    render(<Home />)
    expect(screen.getByText(/get started/i)).toBeInTheDocument()
  })

  it('renders the Deploy Now link', () => {
    render(<Home />)
    expect(screen.getByText('Deploy Now')).toBeInTheDocument()
  })

  it('renders the Documentation link', () => {
    render(<Home />)
    expect(screen.getByText('Documentation')).toBeInTheDocument()
  })
})

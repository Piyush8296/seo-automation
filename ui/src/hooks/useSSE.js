import { useEffect, useRef, useState, useCallback } from 'react'

/**
 * Subscribes to an SSE endpoint and returns the latest event and a done flag.
 * Automatically closes the connection once a terminal event (complete/error/cancelled) arrives.
 */
export function useSSE(url) {
  const [lastEvent, setLastEvent] = useState(null)
  const [done, setDone] = useState(false)
  const esRef = useRef(null)

  const close = useCallback(() => {
    if (esRef.current) {
      esRef.current.close()
      esRef.current = null
    }
  }, [])

  useEffect(() => {
    if (!url) return

    const es = new EventSource(url)
    esRef.current = es

    es.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data)
        setLastEvent(event)
        if (event.type === 'complete' || event.type === 'error' || event.type === 'cancelled') {
          setDone(true)
          es.close()
          esRef.current = null
        }
      } catch {
        // ignore malformed frames
      }
    }

    es.onerror = () => {
      setDone(true)
      es.close()
      esRef.current = null
    }

    return () => {
      es.close()
      esRef.current = null
    }
  }, [url])

  return { lastEvent, done, close }
}

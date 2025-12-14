import React, { useEffect, useState } from 'react'

type Task = {
  id: number
  title: string
  completed: boolean
  created_at: string
}

//@ts-expect-error any
const API_URL = window._env_?.VITE_API_URL || 'http://localhost:8080';

const App: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([])
  const [newTitle, setNewTitle] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchTasks = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch(`${API_URL}/api/tasks`)
      if (!res.ok) throw new Error('Failed to load tasks')
      const data = await res.json()
      setTasks(data)
    } catch (err: any) {
      setError(err.message || 'Error fetching tasks')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTasks()
  }, [])

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTitle.trim()) return
    try {
      const res = await fetch(`${API_URL}/api/tasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: newTitle }),
      })
      if (!res.ok) throw new Error('Failed to create task')
      const created = await res.json()
      setTasks(prev => [created, ...prev])
      setNewTitle('')
    } catch (err: any) {
      setError(err.message || 'Error creating task')
    }
  }

  const toggleCompleted = async (task: Task) => {
    try {
      const res = await fetch(`${API_URL}/api/tasks/${task.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ completed: !task.completed }),
      })
      if (!res.ok) throw new Error('Failed to update task')
      const updated = await res.json()
      setTasks(prev => prev.map(t => (t.id === updated.id ? updated : t)))
    } catch (err: any) {
      setError(err.message || 'Error updating task')
    }
  }

  const deleteTask = async (id: number) => {
    try {
      const res = await fetch(`${API_URL}/api/tasks/${id}`, {
        method: 'DELETE',
      })
      if (!res.ok && res.status !== 204) throw new Error('Failed to delete task')
      setTasks(prev => prev.filter(t => t.id !== id))
    } catch (err: any) {
      setError(err.message || 'Error deleting task')
    }
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'flex-start',
      padding: '2rem',
      background: '#0f172a',
      color: '#f9fafb',
      fontFamily: 'system-ui, -apple-system, BlinkMacSystemFont, sans-serif'
    }}>
      <div style={{
        width: '100%',
        maxWidth: '600px',
        background: '#020617',
        borderRadius: '1rem',
        padding: '1.5rem',
        boxShadow: '0 10px 30px rgba(0,0,0,0.4)'
      }}>
        <h1 style={{ fontSize: '1.7rem', marginBottom: '1rem' }}>TaskBoard 🚀</h1>
        <p style={{ fontSize: '0.9rem', color: '#9ca3af', marginBottom: '1.5rem' }}>
          Demo app (Go + React + PostgreSQL) for CI/CD pipelines.
        </p>

        <form onSubmit={handleAdd} style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
          <input
            type="text"
            placeholder="Add a new task..."
            value={newTitle}
            onChange={e => setNewTitle(e.target.value)}
            style={{
              flex: 1,
              padding: '0.6rem 0.8rem',
              borderRadius: '0.7rem',
              border: '1px solid #1f2937',
              background: '#020617',
              color: 'white',
              outline: 'none'
            }}
          />
          <button
            type="submit"
            style={{
              padding: '0.6rem 1.2rem',
              borderRadius: '0.7rem',
              border: 'none',
              background: '#22c55e',
              color: '#022c22',
              fontWeight: 600,
              cursor: 'pointer'
            }}
          >
            Add
          </button>
        </form>

        {error && (
          <div style={{
            background: '#7f1d1d',
            color: '#fee2e2',
            padding: '0.5rem 0.75rem',
            borderRadius: '0.6rem',
            fontSize: '0.85rem',
            marginBottom: '0.75rem'
          }}>
            {error}
          </div>
        )}

        {loading ? (
          <p>Loading tasks...</p>
        ) : tasks.length === 0 ? (
          <p style={{ color: '#9ca3af', fontSize: '0.9rem' }}>No tasks yet. Create your first one! ✨</p>
        ) : (
          <ul style={{ listStyle: 'none', padding: 0, margin: 0, display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            {tasks.map(task => (
              <li key={task.id} style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '0.6rem 0.8rem',
                borderRadius: '0.7rem',
                background: '#020617',
                border: '1px solid #111827'
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
                  <input
                    type="checkbox"
                    checked={task.completed}
                    onChange={() => toggleCompleted(task)}
                  />
                  <span style={{
                    textDecoration: task.completed ? 'line-through' : 'none',
                    color: task.completed ? '#6b7280' : '#e5e7eb'
                  }}>
                    {task.title}
                  </span>
                </div>
                <button
                  onClick={() => deleteTask(task.id)}
                  style={{
                    border: 'none',
                    background: 'transparent',
                    color: '#fca5a5',
                    cursor: 'pointer',
                    fontSize: '0.85rem'
                  }}
                >
                  Delete
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

export default App

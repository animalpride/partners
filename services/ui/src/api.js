const CORE_BASE = '/api/core/v1'
const AUTH_BASE = '/api/auth/v1'

function getCookie(name) {
  const pair = document.cookie
    .split(';')
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${name}=`))
  if (!pair) return ''
  return decodeURIComponent(pair.slice(name.length + 1))
}

export async function getPage(slug) {
  const response = await fetch(`${CORE_BASE}/cms/pages/${slug}`, { credentials: 'include' })
  if (!response.ok) {
    throw new Error('Failed to fetch page')
  }
  return response.json()
}

export async function updatePage(slug, payload) {
  const response = await fetch(`${CORE_BASE}/cms/admin/pages/${slug}`, {
    method: 'PUT',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Failed to update page')
  }
  return response.json()
}

export async function getCanEditCms() {
  const response = await fetch(`${AUTH_BASE}/permissions`, {
    credentials: 'include',
  })
  if (!response.ok) {
    return false
  }

  const permissions = await response.json()
  return permissions.some((p) => p.resource === 'cms' && p.action === 'edit')
}

export async function getAuthState() {
  const response = await fetch(`${AUTH_BASE}/permissions`, {
    credentials: 'include',
  })

  if (!response.ok) {
    return { authenticated: false, canEditCms: false }
  }

  const permissions = await response.json()
  return {
    authenticated: true,
    canEditCms: permissions.some((p) => p.resource === 'cms' && p.action === 'edit'),
  }
}

export async function login(payload) {
  const response = await fetch(`${AUTH_BASE}/login`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Login failed')
  }

  return response.json()
}

export async function logout() {
  const csrfToken = getCookie('csrf_token')
  const response = await fetch(`${AUTH_BASE}/logout`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'X-CSRF-Token': csrfToken,
    },
  })

  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Logout failed')
  }

  return response.json()
}

export async function submitApplication(payload) {
  const response = await fetch(`${CORE_BASE}/partners/leads`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Failed to submit application')
  }

  return response.json()
}

export async function getComingSoonState() {
  const response = await fetch(`${CORE_BASE}/site/coming-soon`, {
    credentials: 'include',
  })

  if (!response.ok) {
    return {
      enabled: false,
      preview_unlocked: false,
      message: '',
    }
  }

  return response.json()
}

export async function unlockComingSoonPreview(token) {
  const response = await fetch(`${CORE_BASE}/site/coming-soon/unlock/${encodeURIComponent(token)}`, {
    method: 'POST',
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error('Preview unlock failed')
  }

  return response.json()
}

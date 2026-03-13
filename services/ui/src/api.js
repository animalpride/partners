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

// Authenticated fetch wrapper — redirects to / on session expiry (401)
async function authFetch(url, options = {}) {
  const response = await fetch(url, { credentials: 'include', ...options })
  if (response.status === 401) {
    window.location.replace('/')
    throw new Error('Session expired')
  }
  return response
}

export async function getPage(slug) {
  const response = await fetch(`${CORE_BASE}/cms/pages/${slug}`, { credentials: 'include' })
  if (!response.ok) {
    throw new Error('Failed to fetch page')
  }
  return response.json()
}

export async function updatePage(slug, payload) {
  const response = await authFetch(`${CORE_BASE}/cms/admin/pages/${slug}`, {
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

// ── CSRF ──────────────────────────────────────────────────────────────────────

export async function fetchCsrf() {
  await fetch(`${AUTH_BASE}/csrf`, { credentials: 'include' })
}

// ── Invitations (public — no auth required) ───────────────────────────────────

export async function validateInvitation(token) {
  const csrfToken = getCookie('csrf_token')
  const response = await fetch(`${AUTH_BASE}/invitations/validate`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
    },
    body: JSON.stringify({ token }),
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Invalid invitation')
  }
  return response.json()
}

export async function registerInvitation(payload) {
  const csrfToken = getCookie('csrf_token')
  const response = await fetch(`${AUTH_BASE}/invitations/register`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
    },
    body: JSON.stringify(payload),
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Registration failed')
  }
  return response.json()
}

// ── Invitations (admin — requires auth + admin role) ──────────────────────────

export async function getInvitations() {
  const response = await authFetch(`${AUTH_BASE}/invitations/pending`, {
    credentials: 'include',
  })
  if (!response.ok) throw new Error('Failed to fetch invitations')
  return response.json()
}

export async function createInvitation(email, roleId) {
  const csrfToken = getCookie('csrf_token')
  const response = await authFetch(`${AUTH_BASE}/invitations`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
    },
    body: JSON.stringify({ email, role_id: roleId }),
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Failed to send invitation')
  }
  return response.json()
}

export async function resendInvitation(email) {
  const csrfToken = getCookie('csrf_token')
  const response = await authFetch(`${AUTH_BASE}/invitations/resend`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
    },
    body: JSON.stringify({ email }),
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Failed to resend invitation')
  }
  return response.json()
}

export async function revokeInvitation(email) {
  const csrfToken = getCookie('csrf_token')
  const response = await authFetch(`${AUTH_BASE}/invitations/revoke`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
    },
    body: JSON.stringify({ email }),
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Failed to revoke invitation')
  }
  return response.json()
}

export async function getRoles() {
  const response = await authFetch(`${AUTH_BASE}/admin/roles`, {
    credentials: 'include',
  })
  if (!response.ok) throw new Error('Failed to fetch roles')
  return response.json()
}

export async function getUsers() {
  const response = await authFetch(`${AUTH_BASE}/admin/users`, {
    credentials: 'include',
  })
  if (!response.ok) throw new Error('Failed to fetch users')
  return response.json()
}

export async function setUserActive(id, active) {
  const csrfToken = getCookie('csrf_token')
  const response = await authFetch(`${AUTH_BASE}/admin/users/${id}/activate`, {
    method: 'PUT',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
    },
    body: JSON.stringify({ active }),
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Failed to update user status')
  }
  return response.json()
}

export async function assignRole(userId, roleId) {
  const csrfToken = getCookie('csrf_token')
  const response = await authFetch(`${AUTH_BASE}/admin/assign-role`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
    },
    body: JSON.stringify({ user_id: userId, role_id: roleId }),
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Failed to assign role')
  }
  return response.json()
}

export async function removeRole(userId, roleId) {
  const csrfToken = getCookie('csrf_token')
  const response = await authFetch(`${AUTH_BASE}/admin/remove-role`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
    },
    body: JSON.stringify({ user_id: userId, role_id: roleId }),
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.error || 'Failed to remove role')
  }
  return response.json()
}

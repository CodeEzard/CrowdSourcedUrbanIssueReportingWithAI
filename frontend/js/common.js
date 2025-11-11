

const qs = s => document.querySelector(s);
const qsa = s => document.querySelectorAll(s);

// Base URL for backend API when served cross-origin (e.g. frontend on Vercel).
// You can override at runtime by defining window.API_BASE in the HTML before including this script.
// Example in index.html head: <script>window.API_BASE="https://your-backend-domain";</script>
// If not set and we're on a vercel.app domain, provide a placeholder you MUST change.
const API_BASE = (typeof window !== 'undefined' && (window.API_BASE || localStorage.getItem('api_base')))
  || (typeof window !== 'undefined' && window.location.hostname.endsWith('vercel.app')
      ? 'https://REPLACE_WITH_BACKEND_DOMAIN' // TODO: set to your deployed backend host
      : ''); // empty means same-origin

function apiUrl(path) {
  // Ensure single slash joining
  if (!API_BASE) return path; // same-origin
  return API_BASE.replace(/\/$/, '') + path;
}

function apiFetch(path, options = {}) {
  // Add credentials only if we have a cross-origin base and intend to use cookies
  const crossOrigin = API_BASE && !API_BASE.includes(window.location.host);
  const merged = {
    credentials: crossOrigin ? 'include' : 'same-origin',
    ...options
  };
  return fetch(apiUrl(path), merged);
}


const defaultData = [
  { id: 1, title: 'Large pothole on Main Street', category: 'Pothole', location: 'Main Street, Mumbai', desc: 'A deep pothole causing traffic delays.', photo: '', lat: 19.0760, lon: 72.8777, votes: 12, status: 'open', reporter: 'rajesh', comments: [], assignedAt: null },
  { id: 2, title: 'Overflowing trash bin', category: 'Garbage', location: 'Park Avenue, Delhi', desc: 'Trash not collected for over a week.', photo: '', lat: 28.7041, lon: 77.1025, votes: 8, status: 'inprogress', reporter: 'priya', comments: [{ user: 'admin', text: 'Team dispatched.', timestamp: '2025-10-26T14:30:00Z' }], assignedAt: '2025-10-26T14:00:00Z' },
  { id: 3, title: 'Flickering streetlight', category: 'Streetlight', location: 'Colaba, Mumbai', desc: 'Streetlight keeps turning on and off at night.', photo: '', lat: 18.9220, lon: 72.8300, votes: 5, status: 'closed', reporter: 'amit', comments: [{ user: 'admin', text: 'Repaired on 10/25/2025.', timestamp: '2025-10-25T09:15:00Z' }], assignedAt: '2025-10-24T10:00:00Z' },
  { id: 4, title: 'Waterlogging after rain', category: 'Waterlogging', location: 'Andheri East, Mumbai', desc: 'Heavy waterlogging blocking the road.', photo: '', lat: 19.1139, lon: 72.8577, votes: 20, status: 'open', reporter: 'sunita', comments: [], assignedAt: null },
  { id: 5, title: 'Broken bench in park', category: 'Garbage', location: 'Lodhi Garden, Delhi', desc: 'Bench is damaged and unsafe to use.', photo: '', lat: 28.5880, lon: 77.2197, votes: 3, status: 'inprogress', reporter: 'vikram', comments: [], assignedAt: '2025-10-27T09:00:00Z' }
];



// Removed persistent LocalStorage storage of issues.
// Frontend now treats the backend as the source of truth for issues.
function loadData() {
  // intentionally return empty list — issues are fetched from the server
  return [];
}

function saveData(/*data*/) {
  // no-op: we intentionally do not persist issues to localStorage
  // to ensure server is the canonical store in deployed environments.
}

// Fetch feed from backend /feed and map backend posts to frontend issue shape.
async function fetchFeedFromServer() {
  try {
    const resp = await apiFetch('/feed');
    if (!resp.ok) throw new Error('Failed to fetch feed: ' + resp.status);
    const posts = await resp.json();
    if (!Array.isArray(posts)) throw new Error('Invalid feed format');
    const mapped = posts.map(p => {
      // Backend post shape (models.Post) -> frontend issue shape
      const issue = p.issue || {};
      const user = p.user || {};
      const comments = Array.isArray(p.comments) ? p.comments.map(c => ({
        user: c.user?.name || c.user?.email || 'User',
        text: c.content || '',
        timestamp: c.created_at || new Date().toISOString()
      })) : [];
      const upvoteCount = Array.isArray(p.upvotes) ? p.upvotes.length : 0;
      
      return {
        id: p.id || (p.ID || Date.now()),
        title: issue.name || issue.Name || p.title || 'Issue',
        category: issue.category || issue.Category || 'Other',
        location: issue.description || issue.Description || '',
        desc: p.description || p.Description || issue.description || '',
        photo: p.media_url || p.mediaUrl || p.MediaURL || '',
        lat: p.lat || p.Lat || null,
        lon: p.lng || p.Lng || null,
        votes: upvoteCount,
        status: p.status || p.Status || 'open',
        reporter: user.name || user.email || (p.reporter || 'Unknown'),
        comments: comments,
        assignedAt: p.assigned_at || p.assignedAt || null
      };
    });
    // Use server posts as authoritative source. Do not merge with local saved
    // items or persist feed to localStorage in production/deployable mode.
    issues = mapped;
    document.dispatchEvent(new CustomEvent('issuesUpdated'));
    return issues;
  } catch (err) {
    console.warn('fetchFeedFromServer failed:', err);
    return null;
  }
}


let issues = [];
let upvotes = {};
let role = localStorage.getItem('uc_role') || 'user';
let currentUser = localStorage.getItem('uc_user') || 'Guest';
localStorage.setItem('uc_user', currentUser);
let profilePic =
  localStorage.getItem('uc_profile_pic') ||
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48"><circle cx="24" cy="24" r="24" fill="%23e6eef6"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" fill="%23004d40">User</text></svg>';
let geoCache = JSON.parse(localStorage.getItem('uc_geo_cache') || '{}');

// Authentication status management
function updateAuthUI() {
  const jwt = localStorage.getItem('jwt');
  const user = localStorage.getItem('uc_user');
  const loginBtn = document.getElementById('login-btn');
  const logoutBtn = document.getElementById('logout-btn');
  const userDisplay = document.getElementById('user-display');

  if (jwt && user) {
    // User is logged in
    if (loginBtn) loginBtn.style.display = 'none';
    if (logoutBtn) {
      logoutBtn.style.display = 'block';
      logoutBtn.addEventListener('click', logout);
    }
    if (userDisplay) userDisplay.textContent = `👤 ${user}`;
  } else {
    // User is not logged in
    if (loginBtn) loginBtn.style.display = 'block';
    if (logoutBtn) logoutBtn.style.display = 'none';
    if (userDisplay) userDisplay.textContent = '';
  }
}

// Logout function
async function logout() {
  const jwt = localStorage.getItem('jwt');
  if (!jwt) {
    localStorage.removeItem('jwt');
    localStorage.removeItem('uc_user');
    updateAuthUI();
    window.location.href = 'login2.html';
    return;
  }

  try {
    const response = await apiFetch('/logout', {
      method: 'POST',
      headers: { 'Authorization': 'Bearer ' + jwt }
    });

    if (response.ok) {
      console.log('Logout successful');
    } else {
      console.warn('Logout response:', response.status);
    }
  } catch (error) {
    console.error('Logout error:', error);
  }

  // Clear local storage and redirect
  localStorage.removeItem('jwt');
  localStorage.removeItem('uc_user');
  updateAuthUI();
  window.location.href = 'login2.html';
}

// Check if user needs to be logged in for current page
function requireAuth(message = 'You need to be logged in to access this page') {
  const jwt = localStorage.getItem('jwt');
  if (!jwt) {
    alert(message);
    window.location.href = 'login2.html';
    return false;
  }
  return true;
}

// Call this on page load to update auth UI
document.addEventListener('DOMContentLoaded', () => {
  updateAuthUI();
});

let lastDeletedIssue = null;



function showToast(message, type = 'success') {
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.setAttribute('role', 'alert');
  toast.setAttribute('aria-live', 'assertive');
  toast.innerHTML = message;
  document.body.appendChild(toast);
  setTimeout(() => toast.remove(), 3000);
}

function getCategoryIcon(category) {
  const icons = {
    Pothole: '🚧',
    Garbage: '🗑️',
    Streetlight: '💡',
    Waterlogging: '💧'
  };
  return icons[category] || '📍';
}

function observeIssues(container) {
  const observer = new IntersectionObserver(
    entries => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          entry.target.style.opacity = 1;
          entry.target.style.transform = 'translateY(0)';
          observer.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.1 }
  );

  container.querySelectorAll('.issue').forEach(issue => {
    issue.style.opacity = 0;
    issue.style.transform = 'translateY(20px)';
    issue.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
    observer.observe(issue);
  });
}

function debounce(fn, delay) {
  let timeout;
  return (...args) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => fn(...args), delay);
  };
}



async function reverseGeocode(lat, lon) {
  const cacheKey = `${lat},${lon}`;
  if (geoCache[cacheKey]) return geoCache[cacheKey];
  qs('#r-location').classList.add('loading');
  try {
    const response = await fetch(
      `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lon}&zoom=18&addressdetails=1`,
      { headers: { 'User-Agent': 'UrbanCivic/1.0 (Demo Project)' } }
    );
    const data = await response.json();
    if (data?.display_name) {
      geoCache[cacheKey] = data.display_name;
      localStorage.setItem('uc_geo_cache', JSON.stringify(geoCache));
      return data.display_name;
    }
    throw new Error('No address found');
  } catch (error) {
    console.warn('Reverse geocoding failed:', error);
    return 'Unknown location';
  } finally {
    qs('#r-location').classList.remove('loading');
  }
}


// No storage synchronization needed for issues (server-backed)

// Helper: post a comment to the backend. Requires Authorization header (jwt in localStorage).
// Automatically refreshes the feed after posting.
async function postComment(postID, content) {
  const token = localStorage.getItem('jwt');
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = 'Bearer ' + token;
  const resp = await apiFetch('/comment', {
    method: 'POST',
    headers,
    body: JSON.stringify({ post_id: postID, content })
  });
  if (!resp.ok) {
    const text = await resp.text().catch(() => '');
    throw new Error('Comment failed: ' + resp.status + ' ' + text);
  }
  const comment = await resp.json();
  
  // Refresh feed to get updated comments from server
  await fetchFeedFromServer();
  
  return comment;
}

// Helper: toggle upvote for a post. Returns { upvoted: bool }
// Automatically refreshes the feed after toggling upvote.
async function postUpvote(postID) {
  const token = localStorage.getItem('jwt');
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = 'Bearer ' + token;
  const resp = await apiFetch('/upvote', {
    method: 'POST',
    headers,
    body: JSON.stringify({ post_id: postID })
  });
  if (!resp.ok) {
    const text = await resp.text().catch(() => '');
    throw new Error('Upvote failed: ' + resp.status + ' ' + text);
  }
  const result = await resp.json();
  
  // Refresh feed to get updated upvote count from server
  await fetchFeedFromServer();
  
  return result;
}

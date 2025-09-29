// Configuration
const API_BASE_URL = 'http://localhost:8080';
const TOKEN_KEY = 'auth_token';
const USER_KEY = 'user_data';

// Application state
let currentUser = null;
let authToken = null;

/**
 * Initialize the application when the page loads
 */
function initializeApp() {
    showMessage('System initialized and ready for authentication', 'info');
    
    const storedToken = localStorage.getItem(TOKEN_KEY);
    const storedUser = localStorage.getItem(USER_KEY);
    
    if (storedToken && storedUser) {
        try {
            authToken = storedToken;
            currentUser = JSON.parse(storedUser);
            validateStoredSession();
        } catch (error) {
            console.error('Error parsing stored user data:', error);
            clearStoredAuth();
        }
    }
}

/**
 * Validate that a stored session is still valid
 */
async function validateStoredSession() {
    try {
        const response = await fetchWithAuth('/products');
        if (response.ok) {
            showUserDashboard();
            showMessage('Session restored successfully', 'success');
        } else {
            clearStoredAuth();
            showMessage('Session expired - please sign in again', 'info');
        }
    } catch (error) {
        clearStoredAuth();
        showMessage('Unable to verify session - please sign in', 'error');
    }
}

/**
 * Handle login form submission
 */
async function handleLogin(event) {
    event.preventDefault();
    
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;
    
    setLoginButtonLoading(true);
    
    try {
        const response = await fetch(`${API_BASE_URL}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        if (response.ok) {
            const data = await response.json();
            authToken = data.token;
            currentUser = data.user;
            
            localStorage.setItem(TOKEN_KEY, authToken);
            localStorage.setItem(USER_KEY, JSON.stringify(currentUser));
            
            showUserDashboard();
            showMessage(`Welcome back, ${currentUser.name}!`, 'success');
        } else {
            showMessage('Authentication failed - please check your credentials', 'error');
        }
    } catch (error) {
        console.error('Login error:', error);
        showMessage('Connection failed - ensure the backend server is running', 'error');
    } finally {
        setLoginButtonLoading(false);
    }
}

/**
 * Handle registration form submission
 */
async function handleRegister(event) {
    event.preventDefault();
    
    const name = document.getElementById('registerName').value;
    const email = document.getElementById('registerEmail').value;
    const password = document.getElementById('registerPassword').value;
    
    setRegisterButtonLoading(true);
    
    try {
        const response = await fetch(`${API_BASE_URL}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, email, password })
        });

        if (response.ok) {
            const data = await response.json();
            authToken = data.token;
            currentUser = data.user;
            
            localStorage.setItem(TOKEN_KEY, authToken);
            localStorage.setItem(USER_KEY, JSON.stringify(currentUser));
            
            showUserDashboard();
            showMessage(`Account created successfully! Welcome, ${currentUser.name}!`, 'success');
        } else {
            const errorData = await response.json();
            if (errorData.details && Array.isArray(errorData.details)) {
                const errorMessages = errorData.details.map(err => `${err.field}: ${err.message}`).join('<br>');
                showMessage(`Registration failed:<br>${errorMessages}`, 'error');
            } else {
                showMessage(errorData.error || 'Registration failed', 'error');
            }
        }
    } catch (error) {
        console.error('Registration error:', error);
        showMessage('Connection failed - ensure the backend server is running', 'error');
    } finally {
        setRegisterButtonLoading(false);
    }
}

/**
 * UI control functions
 */
function showLoginForm() {
    document.getElementById('loginForm').classList.remove('hidden');
    document.getElementById('registerForm').classList.add('hidden');
    document.getElementById('loginTab').classList.add('active');
    document.getElementById('registerTab').classList.remove('active');
    document.getElementById('tabIndicator').classList.remove('register-active');
}

function showRegisterForm() {
    document.getElementById('loginForm').classList.add('hidden');
    document.getElementById('registerForm').classList.remove('hidden');
    document.getElementById('loginTab').classList.remove('active');
    document.getElementById('registerTab').classList.add('active');
    document.getElementById('tabIndicator').classList.add('register-active');
}

function showUserDashboard() {
    document.getElementById('authSection').classList.add('hidden');
    document.getElementById('userDashboard').classList.remove('hidden');
    
    const userDetailsDiv = document.getElementById('userDetails');
    const rolesHtml = currentUser.roles.map(role => 
        `<span class="role-badge ${role}">${role}</span>`
    ).join('');
    
    userDetailsDiv.innerHTML = `
        <div class="user-detail"><strong>Name:</strong> ${currentUser.name}</div>
        <div class="user-detail"><strong>Email:</strong> ${currentUser.email}</div>
        <div class="user-detail"><strong>Status:</strong> ${currentUser.is_active ? 'Active' : 'Inactive'}</div>
        <div class="user-detail"><strong>Verified:</strong> ${currentUser.email_verified ? 'Yes' : 'No'}</div>
        <div class="user-detail"><strong>Member Since:</strong> ${new Date(currentUser.created_at).toLocaleDateString()}</div>
        <div class="user-detail">
            <strong>Roles:</strong>
            <div class="roles">${rolesHtml}</div>
        </div>
    `;
    
    if (currentUser.roles.includes('admin')) {
        document.getElementById('adminSection').classList.remove('hidden');
    }
}

/**
 * API functions
 */
async function fetchWithAuth(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;
    const authHeaders = {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
        ...options.headers
    };
    return fetch(url, { ...options, headers: authHeaders });
}

async function loadProducts() {
    if (!authToken) {
        showMessage('Authentication required to access products', 'error');
        return;
    }

    try {
        const response = await fetchWithAuth('/products');
        if (response.ok) {
            const products = await response.json();
            displayProducts(products);
            showMessage('Products loaded successfully', 'success');
        } else if (response.status === 401) {
            showMessage('Session expired - please sign in again', 'error');
            logout();
        } else {
            showMessage('Failed to load products', 'error');
        }
    } catch (error) {
        console.error('Error loading products:', error);
        showMessage('Unable to load products - check connection', 'error');
    }
}

async function loadAdminData() {
    try {
        const response = await fetchWithAuth('/admin');
        if (response.ok) {
            const data = await response.json();
            document.getElementById('adminContent').innerHTML = `
                <div style="margin-top: 1rem; padding: 1.5rem; background: rgba(255, 255, 255, 0.05); border-radius: 12px;">
                    <h4>Admin Dashboard</h4>
                    <p><strong>Message:</strong> ${data.message}</p>
                    <p><strong>User:</strong> ${data.user}</p>
                    <p>This content requires administrative privileges to access.</p>
                </div>
            `;
            showMessage('Admin data loaded successfully', 'success');
        } else if (response.status === 403) {
            showMessage('Access denied - insufficient privileges', 'error');
        } else if (response.status === 401) {
            showMessage('Session expired - please sign in again', 'error');
            logout();
        } else {
            showMessage('Failed to load admin data', 'error');
        }
    } catch (error) {
        console.error('Error loading admin data:', error);
        showMessage('Unable to load admin data - check connection', 'error');
    }
}

async function testConnection() {
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        const statusDiv = document.getElementById('connectionStatus');
        
        if (response.ok) {
            statusDiv.innerHTML = '<div class="message success">Backend server is online and responding</div>';
        } else {
            statusDiv.innerHTML = '<div class="message error">Backend server returned an error</div>';
        }
    } catch (error) {
        const statusDiv = document.getElementById('connectionStatus');
        statusDiv.innerHTML = '<div class="message error">Cannot connect to backend server - ensure it\'s running on port 8080</div>';
    }
}

function displayProducts(products) {
    const container = document.getElementById('productsContainer');
    
    if (products.length === 0) {
        container.innerHTML = '<p style="margin-top: 1rem; color: var(--text-secondary);">No products available</p>';
        return;
    }
    
    const productsHtml = products.map(product => `
        <div class="product-card">
            <h4>${product.name}</h4>
            <p style="color: var(--text-secondary); margin-bottom: 1rem;">${product.description || 'No description available'}</p>
            <div class="product-price">$${product.price.toFixed(2)}</div>
            <small style="color: var(--text-muted);">Added: ${new Date(product.created_at).toLocaleDateString()}</small>
        </div>
    `).join('');
    
    container.innerHTML = `<div class="products-grid">${productsHtml}</div>`;
}

/**
 * User management functions
 */
function logout() {
    authToken = null;
    currentUser = null;
    clearStoredAuth();
    
    document.getElementById('userDashboard').classList.add('hidden');
    document.getElementById('adminSection').classList.add('hidden');
    document.getElementById('authSection').classList.remove('hidden');
    
    document.getElementById('loginFormElement').reset();
    document.getElementById('registerFormElement').reset();
    document.getElementById('productsContainer').innerHTML = '';
    document.getElementById('adminContent').innerHTML = '';
    
    showMessage('Successfully signed out', 'info');
}

function clearStoredAuth() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
}

/**
 * Utility functions
 */
function showMessage(message, type = 'info') {
    const messageArea = document.getElementById('messageArea');
    messageArea.innerHTML = `<div class="message ${type}">${message}</div>`;
    
    if (type === 'success' || type === 'info') {
        setTimeout(() => messageArea.innerHTML = '', 5000);
    }
}

function setLoginButtonLoading(loading) {
    const btn = document.getElementById('loginBtn');
    const text = document.getElementById('loginBtnText');
    const spinner = document.getElementById('loginSpinner');
    
    btn.disabled = loading;
    text.textContent = loading ? 'Signing In...' : 'Sign In';
    spinner.classList.toggle('hidden', !loading);
}

function setRegisterButtonLoading(loading) {
    const btn = document.getElementById('registerBtn');
    const text = document.getElementById('registerBtnText');
    const spinner = document.getElementById('registerSpinner');
    
    btn.disabled = loading;
    text.textContent = loading ? 'Creating Account...' : 'Create Account';
    spinner.classList.toggle('hidden', !loading);
}

/**
 * Event listeners - Set up when the DOM is loaded
 */
document.addEventListener('DOMContentLoaded', function() {
    // Initialize the application
    initializeApp();
    
    // Set up form event listeners
    document.getElementById('loginFormElement').addEventListener('submit', handleLogin);
    document.getElementById('registerFormElement').addEventListener('submit', handleRegister);
});

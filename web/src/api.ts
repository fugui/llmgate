import axios from 'axios';

// `baseURL` is left empty so that requests are relative to the current host.
// During development the Vite dev server proxies `/api` & `/v1` to the backend
// defined in `vite.config.ts`. In production you can set an explicit
// `baseURL` via environment variables if the backend lives elsewhere.

const api = axios.create({
  baseURL: '',
});

// Add request interceptor to dynamically set Authorization header
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Add response interceptor to handle 401 errors globally
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear authentication data
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      // Redirect to login page
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;

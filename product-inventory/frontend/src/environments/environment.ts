const isLocal = ['localhost', '127.0.0.1'].includes(window.location.hostname);

export const environment = {
  apiUrl: isLocal
    ? 'http://localhost:8080/api'
    : 'https://product-inventory-management-system-5v8p.onrender.com/api',
};
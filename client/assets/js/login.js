document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('loginForm');
    const togglePassword = document.getElementById('togglePassword');
    const passwordInput = document.getElementById('password');
    const emailInput = document.getElementById('email');

    // Toggle password visibility
    togglePassword.addEventListener('click', function() {
        const type = passwordInput.getAttribute('type') === 'password' ? 'text' : 'password';
        passwordInput.setAttribute('type', type);
        
        // Change icon
        const icon = this.querySelector('svg path');
        if (type === 'text') {
            // Hide icon
            icon.setAttribute('d', 'M11.83,9L15,12.16C15,12.11 15,12.05 15,12A3,3 0 0,0 12,9C11.94,9 11.89,9 11.83,9M7.53,9.8L9.08,11.35C9.03,11.56 9,11.77 9,12A3,3 0 0,0 12,15C12.22,15 12.44,14.97 12.65,14.92L14.2,16.47C13.53,16.8 12.79,17 12,17A5,5 0 0,1 7,12C7,11.21 7.2,10.47 7.53,9.8M2,4.27L4.28,6.55L4.73,7C3.08,8.3 1.78,10 1,12C2.73,16.39 7,19.5 12,19.5C13.55,19.5 15.03,19.2 16.38,18.66L16.81,19.09L19.73,22L21,20.73L3.27,3M12,7A5,5 0 0,1 17,12C17,12.64 16.87,13.26 16.64,13.82L19.57,16.75C21.07,15.5 22.27,13.86 23,12C21.27,7.61 17,4.5 12,4.5C10.6,4.5 9.26,4.75 8,5.2L10.17,7.35C10.76,7.13 11.37,7 12,7Z');
        } else {
            // Show icon
            icon.setAttribute('d', 'M12,9A3,3 0 0,0 9,12A3,3 0 0,0 12,15A3,3 0 0,0 15,12A3,3 0 0,0 12,9M12,17A5,5 0 0,1 7,12A5,5 0 0,1 12,7A5,5 0 0,1 17,12A5,5 0 0,1 12,17M12,4.5C7,4.5 2.73,7.61 1,12C2.73,16.39 7,19.5 12,19.5C17,19.5 21.27,16.39 23,12C21.27,7.61 17,4.5 12,4.5Z');
        }
    });

    // Form validation and submission
    loginForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        const email = emailInput.value.trim();
        const password = passwordInput.value;
        const remember = document.getElementById('remember').checked;
        
        // Clear previous messages
        clearMessages();
        
        // Basic validation
        if (!email || !password) {
            showError('Vui lòng nhập đầy đủ thông tin');
            return;
        }
        
        if (!isValidEmail(email)) {
            showError('Email không hợp lệ');
            return;
        }
        
        if (password.length < 6) {
            showError('Mật khẩu phải có ít nhất 6 ký tự');
            return;
        }
        
        // Show loading state
        const button = this.querySelector('.login-button');
        setLoadingState(button, true);
        
        // Simulate login API call
        performLogin(email, password, remember)
            .then(response => {
                if (response.success) {
                    // Store user session
                    const userData = {
                        email: email,
                        name: response.user.name || 'User',
                        token: response.token,
                        loginTime: new Date().toISOString()
                    };
                    
                    if (remember) {
                        localStorage.setItem('user', JSON.stringify(userData));
                    } else {
                        sessionStorage.setItem('user', JSON.stringify(userData));
                    }
                    
                    showSuccess('Đăng nhập thành công! Đang chuyển hướng...');
                    
                    // Redirect to chat page after short delay
                    setTimeout(() => {
                        window.location.href = 'index.html';
                    }, 1500);
                } else {
                    throw new Error(response.message || 'Đăng nhập thất bại');
                }
            })
            .catch(error => {
                showError(error.message);
                setLoadingState(button, false);
            });
    });

    // Social login handlers
    document.querySelector('.social-button.google').addEventListener('click', function() {
        handleSocialLogin('google');
    });

    document.querySelector('.social-button.facebook').addEventListener('click', function() {
        handleSocialLogin('facebook');
    });

    // Auto-fill form for demo purposes
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
        // Demo credentials for development
        emailInput.value = 'demo@example.com';
        passwordInput.value = 'password123';
    }

    // Helper functions
    function performLogin(email, password, remember) {
        return new Promise((resolve, reject) => {
            // Simulate API delay
            setTimeout(() => {
                // Demo authentication - replace with actual API call
                if (email === 'demo@example.com' && password === 'password123') {
                    resolve({
                        success: true,
                        user: {
                            name: 'Demo User',
                            email: email
                        },
                        token: 'demo-jwt-token-' + Date.now()
                    });
                } else if (email === 'admin@chat.com' && password === 'admin123') {
                    resolve({
                        success: true,
                        user: {
                            name: 'Admin',
                            email: email
                        },
                        token: 'admin-jwt-token-' + Date.now()
                    });
                } else {
                    reject(new Error('Email hoặc mật khẩu không chính xác'));
                }
            }, 1500);
        });
    }

    function handleSocialLogin(provider) {
        showError(`Tính năng đăng nhập ${provider === 'google' ? 'Google' : 'Facebook'} đang được phát triển`);
        
        // In real implementation, you would integrate with:
        // - Google OAuth 2.0 for Google login
        // - Facebook SDK for Facebook login
        console.log(`${provider} login clicked`);
    }

    function isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    function setLoadingState(button, loading) {
        if (loading) {
            button.classList.add('loading');
            button.disabled = true;
            button.textContent = 'Đang đăng nhập...';
        } else {
            button.classList.remove('loading');
            button.disabled = false;
            button.innerHTML = `
                Đăng nhập
                <svg width="16" height="16" viewBox="0 0 24 24" fill="white">
                    <path d="M8.59,16.58L13.17,12L8.59,7.41L10,6L16,12L10,18L8.59,16.58Z"/>
                </svg>
            `;
        }
    }

    function showError(message) {
        clearMessages();
        const errorDiv = createMessage(message, 'error');
        insertMessage(errorDiv);
    }

    function showSuccess(message) {
        clearMessages();
        const successDiv = createMessage(message, 'success');
        insertMessage(successDiv);
    }

    function createMessage(message, type) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `${type}-message`;
        messageDiv.textContent = message;
        return messageDiv;
    }

    function insertMessage(messageElement) {
        const loginButton = document.querySelector('.login-button');
        loginButton.parentNode.insertBefore(messageElement, loginButton);

        // Auto remove after 5 seconds
        setTimeout(() => {
            if (messageElement.parentNode) {
                messageElement.remove();
            }
        }, 5000);
    }

    function clearMessages() {
        const existingMessages = document.querySelectorAll('.error-message, .success-message');
        existingMessages.forEach(message => message.remove());
    }

    // Check if already logged in
    function checkExistingLogin() {
        const userData = localStorage.getItem('user') || sessionStorage.getItem('user');
        if (userData) {
            try {
                const user = JSON.parse(userData);
                // Verify token is still valid (in real app, check with server)
                if (user.token) {
                    window.location.href = 'index.html';
                }
            } catch (e) {
                // Invalid session data, clear it
                localStorage.removeItem('user');
                sessionStorage.removeItem('user');
            }
        }
    }

    // Input validation feedback
    function setupInputValidation() {
        emailInput.addEventListener('blur', function() {
            if (this.value && !isValidEmail(this.value)) {
                this.style.borderColor = '#DC2626';
            } else {
                this.style.borderColor = '#D1D5DB';
            }
        });

        passwordInput.addEventListener('input', function() {
            if (this.value && this.value.length < 6) {
                this.style.borderColor = '#DC2626';
            } else {
                this.style.borderColor = '#D1D5DB';
            }
        });
    }

    // Initialize
    checkExistingLogin();
    setupInputValidation();
});
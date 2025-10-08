document.addEventListener('DOMContentLoaded', function() {
    const registerForm = document.getElementById('registerForm');
    const fullNameInput = document.getElementById('fullName');
    const emailInput = document.getElementById('email');
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirmPassword');
    const agreeTermsInput = document.getElementById('agreeTerms');

    // Form validation and submission
    registerForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        const fullName = fullNameInput.value.trim();
        const email = emailInput.value.trim();
        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;
        const agreeTerms = agreeTermsInput.checked;
        
        // Clear previous messages
        clearMessages();
        clearFieldErrors();
        
        // Validation
        let isValid = true;
        
        if (!fullName) {
            showFieldError(fullNameInput, 'Vui lòng nhập họ và tên');
            isValid = false;
        } else if (fullName.length < 2) {
            showFieldError(fullNameInput, 'Họ tên phải có ít nhất 2 ký tự');
            isValid = false;
        } else {
            showFieldSuccess(fullNameInput, 'Hợp lệ');
        }
        
        if (!email) {
            showFieldError(emailInput, 'Vui lòng nhập email');
            isValid = false;
        } else if (!isValidEmail(email)) {
            showFieldError(emailInput, 'Email không hợp lệ');
            isValid = false;
        } else {
            showFieldSuccess(emailInput, 'Email hợp lệ');
        }
        
        if (!password) {
            showFieldError(passwordInput, 'Vui lòng nhập mật khẩu');
            isValid = false;
        } else if (password.length < 6) {
            showFieldError(passwordInput, 'Mật khẩu phải có ít nhất 6 ký tự');
            isValid = false;
        } else {
            showFieldSuccess(passwordInput, 'Mật khẩu hợp lệ');
        }
        
        if (!confirmPassword) {
            showFieldError(confirmPasswordInput, 'Vui lòng xác nhận mật khẩu');
            isValid = false;
        } else if (password !== confirmPassword) {
            showFieldError(confirmPasswordInput, 'Mật khẩu xác nhận không khớp');
            isValid = false;
        } else {
            showFieldSuccess(confirmPasswordInput, 'Mật khẩu khớp');
        }
        
        if (!agreeTerms) {
            showError('Bạn phải đồng ý với điều khoản sử dụng');
            isValid = false;
        }
        
        if (!isValid) return;
        
        // Show loading state
        const button = this.querySelector('.register-button');
        setLoadingState(button, true);
        
        // Simulate registration API call
        performRegistration(fullName, email, password)
            .then(response => {
                if (response.success) {
                    showSuccess('Đăng ký thành công! Đang chuyển đến trang đăng nhập...');
                    
                    // Redirect to login page after delay
                    setTimeout(() => {
                        window.location.href = 'login.html?registered=true&email=' + encodeURIComponent(email);
                    }, 2000);
                } else {
                    throw new Error(response.message || 'Đăng ký thất bại');
                }
            })
            .catch(error => {
                showError(error.message);
                setLoadingState(button, false);
            });
    });

    // Real-time validation
    setupRealTimeValidation();

    // Password strength indicator
    setupPasswordStrength();

    // Helper functions
    function performRegistration(fullName, email, password) {
        return new Promise((resolve, reject) => {
            // Simulate API delay
            setTimeout(() => {
                // Demo registration - replace with actual API call
                const existingEmails = ['test@example.com', 'admin@example.com'];
                
                if (existingEmails.includes(email.toLowerCase())) {
                    reject(new Error('Email này đã được sử dụng'));
                } else {
                    resolve({
                        success: true,
                        user: {
                            name: fullName,
                            email: email
                        }
                    });
                }
            }, 2000);
        });
    }

    function setupRealTimeValidation() {
        // Full name validation
        fullNameInput.addEventListener('blur', function() {
            const value = this.value.trim();
            if (value) {
                if (value.length < 2) {
                    showFieldError(this, 'Họ tên phải có ít nhất 2 ký tự');
                } else {
                    showFieldSuccess(this, 'Hợp lệ');
                }
            }
        });

        // Email validation
        emailInput.addEventListener('blur', function() {
            const value = this.value.trim();
            if (value) {
                if (!isValidEmail(value)) {
                    showFieldError(this, 'Email không hợp lệ');
                } else {
                    showFieldSuccess(this, 'Email hợp lệ');
                    // Check if email exists (in real app)
                    checkEmailExists(value);
                }
            }
        });

        // Password validation
        passwordInput.addEventListener('input', function() {
            updatePasswordStrength(this.value);
            
            // Revalidate confirm password if it has value
            if (confirmPasswordInput.value) {
                validateConfirmPassword();
            }
        });

        // Confirm password validation
        confirmPasswordInput.addEventListener('input', validateConfirmPassword);
    }

    function setupPasswordStrength() {
        const passwordGroup = passwordInput.parentNode;
        const strengthIndicator = document.createElement('div');
        strengthIndicator.className = 'password-strength';
        strengthIndicator.innerHTML = `
            <div class="strength-bar">
                <div class="strength-fill"></div>
            </div>
            <div class="strength-text">Độ mạnh mật khẩu</div>
        `;
        passwordGroup.appendChild(strengthIndicator);
    }

    function updatePasswordStrength(password) {
        const strengthFill = document.querySelector('.strength-fill');
        const strengthText = document.querySelector('.strength-text');
        
        if (!password) {
            strengthFill.className = 'strength-fill';
            strengthText.textContent = 'Độ mạnh mật khẩu';
            return;
        }
        
        let strength = 0;
        
        // Length
        if (password.length >= 6) strength++;
        if (password.length >= 8) strength++;
        
        // Contains uppercase
        if (/[A-Z]/.test(password)) strength++;
        
        // Contains lowercase
        if (/[a-z]/.test(password)) strength++;
        
        // Contains numbers
        if (/\d/.test(password)) strength++;
        
        // Contains special characters
        if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) strength++;
        
        if (strength <= 2) {
            strengthFill.className = 'strength-fill weak';
            strengthText.textContent = 'Yếu';
        } else if (strength <= 4) {
            strengthFill.className = 'strength-fill medium';
            strengthText.textContent = 'Trung bình';
        } else {
            strengthFill.className = 'strength-fill strong';
            strengthText.textContent = 'Mạnh';
        }
    }

    function validateConfirmPassword() {
        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;
        
        if (confirmPassword) {
            if (password !== confirmPassword) {
                showFieldError(confirmPasswordInput, 'Mật khẩu không khớp');
            } else {
                showFieldSuccess(confirmPasswordInput, 'Mật khẩu khớp');
            }
        }
    }

    function checkEmailExists(email) {
        // In real app, make API call to check if email exists
        const existingEmails = ['test@example.com', 'admin@example.com'];
        
        if (existingEmails.includes(email.toLowerCase())) {
            showFieldError(emailInput, 'Email này đã được sử dụng');
        }
    }

    function isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    function setLoadingState(button, loading) {
        if (loading) {
            button.classList.add('loading');
            button.disabled = true;
            button.textContent = 'Đang đăng ký...';
        } else {
            button.classList.remove('loading');
            button.disabled = false;
            button.innerHTML = `
                Đăng ký
                <svg width="16" height="16" viewBox="0 0 24 24" fill="white">
                    <path d="M8.59,16.58L13.17,12L8.59,7.41L10,6L16,12L10,18L8.59,16.58Z"/>
                </svg>
            `;
        }
    }

    function showFieldError(input, message) {
        input.classList.remove('success');
        input.classList.add('error');
        
        // Remove existing error message
        const existingError = input.parentNode.querySelector('.field-error');
        if (existingError) existingError.remove();
        
        // Add error message
        const errorDiv = document.createElement('div');
        errorDiv.className = 'field-error';
        errorDiv.textContent = message;
        input.parentNode.appendChild(errorDiv);
    }

    function showFieldSuccess(input, message) {
        input.classList.remove('error');
        input.classList.add('success');
        
        // Remove existing messages
        const existingError = input.parentNode.querySelector('.field-error');
        const existingSuccess = input.parentNode.querySelector('.field-success');
        if (existingError) existingError.remove();
        if (existingSuccess) existingSuccess.remove();
        
        // Add success message
        const successDiv = document.createElement('div');
        successDiv.className = 'field-success';
        successDiv.textContent = message;
        input.parentNode.appendChild(successDiv);
    }

    function clearFieldErrors() {
        document.querySelectorAll('.field-error, .field-success').forEach(el => el.remove());
        document.querySelectorAll('input').forEach(input => {
            input.classList.remove('error', 'success');
        });
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
        const registerButton = document.querySelector('.register-button');
        registerButton.parentNode.insertBefore(messageElement, registerButton);

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

    // Initialize
    checkExistingLogin();
});
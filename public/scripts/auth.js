// DOM Elements
const loginForm = document.getElementById("login-form");
const forgotPassForm = document.getElementById("forgot-password-form");
const otpForm = document.getElementById("otp-form");
const resetPassForm = document.getElementById("reset-pass-form");
const buttonLogin = document.getElementById("btn-login");
const buttonForgot = document.getElementById("btn-forgot");
const buttonSubmitOtp = document.getElementById("btn-submit-otp");
const buttonResetPass = document.getElementById("btn-reset-pass");
const titleAlert = document.getElementById("title-alert");
const resendOtpButton = document.getElementById('resendOtpButton');
const resendTimerSpan = document.getElementById('resendTimer');

const Toast = Swal.mixin({
  toast: true,
  position: "top-end",
  showConfirmButton: false,
  timer: 3000,
  timerProgressBar: true,
  didOpen: (toast) => {
    toast.onmouseenter = Swal.stopTimer;
    toast.onmouseleave = Swal.resumeTimer;
  },
});

// Variables
let email = "";
let otp = "";
let resendTimer;


// Helper Functions
function showElement(id) {
  document.getElementById(id).style.display = "block";
}

function hideElement(id) {
  document.getElementById(id).style.display = "none";
}

function showErrorToast(message) {
  Toast.fire({
      icon: 'error',
      title: message,
  });
}

function showSuccessToast(message) {
  Toast.fire({
      icon: 'success',
      title: message,
  });
}


function isFormDataEmpty(formData) {
  for (let pair of formData.entries()) {
    return false; // If there's at least one entry, the FormData is not empty
  }
  return true; // If we've gone through all entries and found none, the FormData is empty
}

// Form Submission Function
async function submitForm(
  formData,
  url,
  button,
  successCallback,
  errorCallback
) {
  // Log FormData contents
  console.log("Form Data Contents:");
  for (let [key, value] of formData.entries()) {
    console.log(key, value);
  }

  // Check if FormData is empty
  if (isFormDataEmpty(formData)) {
    console.log("FormData is empty");
    errorCallback(new Error("Form is empty"));
    return;
  }

  setLoadingState(button, true);

  try {
    const response = await fetch(url, {
      method: "POST",
      body: formData,
    });

    const data = await response.json();

    if (!response.ok) {
      throw data;
    }

    successCallback(data);
  } catch (error) {
    console.error("Error:", error);
    errorCallback(error);
  } finally {
    setLoadingState(button, false);
  }
}

// Add this new function for the resend timer
function startResendTimer() {
  let timeLeft = 30;
  setLoadingState(resendOtpButton, false); // Reset loading state
  resendOtpButton.disabled = true;
  
  resendTimer = setInterval(() => {
      if (timeLeft <= 0) {
          clearInterval(resendTimer);
          resendOtpButton.disabled = false;
          resendOtpButton.textContent = 'Resend OTP';
      } else {
          resendOtpButton.textContent = `Resend OTP in ${timeLeft} seconds`;
          timeLeft--;
      }
  }, 1000);
}

// Modify the existing setLoadingState function
function setLoadingState(button, isLoading) {
  button.disabled = isLoading;
  if (isLoading) {
      button.dataset.originalText = button.textContent;
      button.innerHTML = '<div class="spinner-border spinner-border-sm" role="status"><span class="visually-hidden">Loading...</span></div> Loading...';
  } else {
      button.textContent = button.dataset.originalText || button.textContent;
  }
}

// Event Listeners
loginForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const formData = new FormData(loginForm);
  submitForm(
    formData,
    "auth/login",
    buttonLogin,
    (data) => {
      localStorage.setItem("token", data.token);
      window.location.href =
        "/?alert=Login successful&message=Login Successful";
    },
    (error) =>
      showErrorToast(
        error.message || "Invalid Email or password. Please try again."
      )
  );
});

forgotPassForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  email = document.getElementById("forgotEmail").value; // Store the email
  const formData = new FormData(forgotPassForm);
  submitForm(
    formData,
    "forgot-password",
    buttonForgot,
    () => {
      showOtpForm(); // This will now start the resend timer
      forgotPassForm.reset();
      showSuccessToast("OTP sent successfully. Please check your email.");
    },
    (error) =>
      showErrorToast(error.message || "Error sending OTP. Please try again.")
  );
});

otpForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  otp = Array.from(document.querySelectorAll(".otp-input"))
    .map((input) => input.value)
    .join("");
  const formData = new FormData();
  formData.append("email", email);
  formData.append("otp", otp);
  submitForm(
    formData,
    "validate-otp",
    buttonSubmitOtp,
    () => {
      showResetPassForm();
      otpForm.reset();
      showSuccessToast("OTP validated successfully. Please reset your password.");
    },
    (error) => showErrorToast(error.message || "Invalid OTP. Please try again.")
  );
});

resetPassForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const formData = new FormData(resetPassForm);
  formData.append("email", email);
  formData.append("otp", otp);
  submitForm(
    formData,
    "reset-password",
    buttonResetPass,
    () => {
      showLoginForm();
      showSuccessToast("Password reset successfully. Please login with your new password.");
      resetPassForm.reset();
      // Clear the stored email and OTP after successful password reset
      email = "";
      otp = "";
    },
    (error) =>
      showErrorToast(
        error.message || "Error resetting password. Please try again."
      )
  );
});

resendOtpButton.addEventListener('click', async () => {
  const formData = new FormData();
  formData.append('email', email);

  try {
      setLoadingState(resendOtpButton, true);
      const response = await fetch('forgot-password', {
          method: 'POST',
          body: formData
      });

      const data = await response.json();

      if (!response.ok) {
          throw data;
      }

      showSuccessToast('OTP resent successfully. Please check your email.');
      startResendTimer(); // Restart the timer after successful resend
  } catch (error) {
      showErrorToast(error.message || 'Error resending OTP. Please try again.');
      // Reset the button state immediately in case of error
      setLoadingState(resendOtpButton, false);
  }
});

// OTP Input Handling
const otpInputs = document.querySelectorAll(".otp-input");

otpInputs.forEach((input, index) => {
  input.addEventListener("input", (e) => handleOtpInput(e, index));
  input.addEventListener("keydown", (e) => handleOtpKeydown(e, index));
  input.addEventListener("paste", handleOtpPaste);
});

function handleOtpInput(e, index) {
  if (e.inputType === "insertFromPaste") {
    e.preventDefault();
    const pastedData = e.clipboardData
      ? e.clipboardData.getData("text")
      : window.clipboardData.getData("text");
    fillOtpFromPaste(pastedData);
  } else if (e.target.value.length === 1 && index < otpInputs.length - 1) {
    otpInputs[index + 1].focus();
  }
}

function handleOtpKeydown(e, index) {
  if (e.key === "Backspace" && e.target.value === "" && index > 0) {
    otpInputs[index - 1].focus();
  }
}

function handleOtpPaste(e) {
  e.preventDefault();
  const pastedData = e.clipboardData
    ? e.clipboardData.getData("text")
    : window.clipboardData.getData("text");
  fillOtpFromPaste(pastedData);
}

function fillOtpFromPaste(pastedData) {
  const otpDigits = pastedData.replace(/\D/g, "").slice(0, 6).split("");
  otpInputs.forEach((input, index) => {
    input.value = otpDigits[index] || "";
  });
  if (otpDigits.length < 6) {
    otpInputs[Math.min(otpDigits.length, 5)].focus();
  }
}

// Navigation Functions
function showOtpForm() {
  hideElement("forgotPasswordForm");
  showElement("OtpForm");
  document.getElementById(
    "otpSubtitle"
  ).innerHTML = `We've sent a code to <strong>${email}</strong>`;
  startResendTimer(); // Start the timer when showing the OTP form
}

function showResetPassForm() {
  hideElement("OtpForm");
  showElement("ResetPassForm");
}

function showForgotPassword() {
  hideElement("loginForm");
  showElement("forgotPasswordForm");
  hideElement("OtpForm");
}

function showLoginForm() {
  showElement("loginForm");
  hideElement("forgotPasswordForm");
  hideElement("OtpForm");
  hideElement("ResetPassForm");
}

function showForgotPassForm() {
  hideElement("OtpForm");
  showElement("forgotPasswordForm");
}

// Password Toggle Functionality
function setupPasswordToggle() {
  const passwordInput = document.getElementById("password");
  const toggleButton = document.getElementById("togglePassword");
  const toggleIcon = document.getElementById("toggleIcon");

  toggleButton.addEventListener("click", function () {
    const type =
      passwordInput.getAttribute("type") === "password" ? "text" : "password";
    passwordInput.setAttribute("type", type);
    toggleIcon.classList.toggle("bi-eye");
    toggleIcon.classList.toggle("bi-eye-slash");
  });
}

// Initialize
document.addEventListener("DOMContentLoaded", setupPasswordToggle);

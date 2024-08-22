const loginForm = document.getElementById("login-form");
const buttonLogin = document.getElementById("btn-login");
const titleAlert = document.getElementById("title-alert");

loginForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  event.stopPropagation();
    if (loginForm.checkValidity()) {
      const form = new FormData(loginForm);

      const originalText = buttonLogin.textContent;
      buttonLogin.innerHTML = '<div class="spinner-border" role="status"><span class="visually-hidden">Loading...</span></div>';

      fetch("auth/login", {
        method: "POST",
        body: form,
      })
        .then(response => {
          if (!response.ok) {
              return response.json().then(err => { throw err; });
          }
          return response.json();
        })
        .then((data) => {
          buttonLogin.innerHTML = originalText;
          localStorage.setItem("token", data.token);

          window.location.href = "/?alert=Login successful&message=Login Successfull";
            
             // Example of saving JWT token
        })
        .catch((error) => {
          console.error("Error:", error);
          buttonLogin.innerHTML = originalText;
          const errorModal = new bootstrap.Modal(document.getElementById('errorModal'));
          titleAlert.innerHTML = `Invalid Email or password. Please try again.`;
        errorModal.show();
          
        });
    }

}
);

const forgotPassForm = document.getElementById("forgot-password-form");
const buttonForgot = document.getElementById("btn-forgot");
let email;

function showOtpForm() {
    document.getElementById('forgotPasswordForm').style.display = 'none';
    document.getElementById('OtpForm').style.display = 'block';
    email = document.getElementById('forgotEmail').value;
    document.getElementById('otpSubtitle').innerHTML = `We've sent a code to <strong>${email}</strong>`;
}

forgotPassForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  event.stopPropagation();
    if (forgotPassForm.checkValidity()) {
      const form = new FormData(forgotPassForm);

      const originalText = buttonForgot.textContent;
      buttonForgot.innerHTML = '<div class="spinner-border" role="status"><span class="visually-hidden">Loading...</span></div>';

      fetch("forgot-password", {
        method: "POST",
        body: form,
      })
        .then(response => {
          if (!response.ok) {
              return response.json().then(err => { throw err; });
          }
          return response.json();
        })
        .then((data) => {
          buttonForgot.innerHTML = originalText;
          showOtpForm();
          forgotPassForm.reset();
        })
        .catch((error) => {
          console.error("Error:", error);
          buttonForgot.innerHTML = originalText;
          const errorModal = new bootstrap.Modal(document.getElementById('errorModal'));
          titleAlert.innerHTML = `Error `;
        errorModal.show();
          
        });
    }

}
);

const otpForm = document.getElementById("otp-form");
const buttonSubmitOtp = document.getElementById("btn-submit-otp");
let otp = "";

function showResetPassForm() {
    document.getElementById('OtpForm').style.display = 'none';
    document.getElementById('ResetPassForm').style.display = 'block';
}

otpForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  event.stopPropagation();
    if (otpForm.checkValidity()) {
      

            for (let i = 1; i <= 6; i++) {
                const otpInput = document.getElementById("otp" + i);
                otp += otpInput.value;
            }
      const form = new FormData(otpForm);
      form.append('email',email);
      form.append('otp',otp);

      const originalText = buttonSubmitOtp.textContent;
      buttonSubmitOtp.innerHTML = '<div class="spinner-border" role="status"><span class="visually-hidden">Loading...</span></div>';

      fetch("validate-otp", {
        method: "POST",
        body: form,
      })
        .then(response => {
          if (!response.ok) {
              return response.json().then(err => { throw err; });
          }
          return response.json();
        })
        .then((data) => {
          buttonSubmitOtp.innerHTML = originalText;
          showResetPassForm();
          otpForm.reset();
        })
        .catch((error) => {
          console.error("Error:", error);
          buttonSubmitOtp.innerHTML = originalText;
          const errorModal = new bootstrap.Modal(document.getElementById('errorModal'));
          titleAlert.innerHTML = `Invalid OTP`;
        errorModal.show();
          
        });
    }

}
);

const resetPassForm = document.getElementById("reset-pass-form");
const buttonResetPass = document.getElementById("btn-reset-pass");

function showLoginForm() {
  document.getElementById('ResetPassForm').style.display = 'none';
  document.getElementById('loginForm').style.display = 'block';
}

resetPassForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  event.stopPropagation();
    if (resetPassForm.checkValidity()) {
      
      const form = new FormData(resetPassForm);
      form.append('email',email);
      form.append('otp',otp);

      const originalText = buttonResetPass.textContent;
      buttonResetPass.innerHTML = '<div class="spinner-border" role="status"><span class="visually-hidden">Loading...</span></div>';

      fetch("reset-password", {
        method: "POST",
        body: form,
      })
        .then(response => {
          if (!response.ok) {
              return response.json().then(err => { throw err; });
          }
          return response.json();
        })
        .then((data) => {
          buttonResetPass.innerHTML = originalText;
          showLoginForm();
          const successModal = new bootstrap.Modal(document.getElementById('successModal'));
          successModal.show();
          resetPassForm.reset();

        })
        .catch((error) => {
          console.error("Error:", error);
          buttonResetPass.innerHTML = originalText;
          const errorModal = new bootstrap.Modal(document.getElementById('errorModal'));
          titleAlert.innerHTML = error.message;
        errorModal.show();
          
        });
    }

}
);


const btnLogout = document.getElementById("btn-logout");

btnLogout.addEventListener("click", async (event) => {
  localStorage.clear();
  window.location.href = "/login";
  console.log("success");
  
} );


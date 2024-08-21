const loginForm = document.getElementById("login-form");
const buttonLogin = document.getElementById("btn-login");

loginForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  event.stopPropagation();
    console.log("asdasd");
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
          console.log(data);
          buttonLogin.innerHTML = 'Login';
            // alert("Login successful!");
            
            window.location.href = "/";
            // Save token or perform other actions like redirecting
            console.log(data.token);
            
            localStorage.setItem("token", data.token); // Example of saving JWT token
        })
        .catch((error) => {
          console.error("Error:", error);
          buttonLogin.innerHTML = 'Login';
          // alert("Error!");
          
        });
    }

  
 
}
);

// async function login(email, password) {
//     const url = "https://api.binav-avts.id:5000/api/auth/login";
//     const data = { email, password };

//     const response = await fetch(url, {
//       method: "POST",
//       headers: { "Content-Type": "application/json" },
//       body: JSON.stringify(data),
//     });

//     if (!response.ok) {
//       throw new Error(`Login failed: ${response.status}`);
//     }

//     return await response.json();
//   }

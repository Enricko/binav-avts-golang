var btnLogin = document.getElementById("btn-login");


// function login(event) {
//     event.preventDefault();
//     window.location.href = "maps.html";
// }

// const loginForm = document.getElementById("login-form");

// loginForm.addEventListener("submit",async(event) =>{
//     event.preventDefault();
//     const email = document.getElementById("email").value;
//     const password = document.getElementById("password").value;


// try {
//     const response = await login(email, password);
//     if (response.status == 200) {

//     window.location.href = "maps.html";
//     }
//   } catch (error) {
//     console.error("Login error:", error);
    
//     messageElement.textContent = "Login failed: " + error.message;
//   }

// } );


// async function login(email, password) {
//     const url = "https://api.binav-avts.id:5000/api/login"; 
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


  const myModal = document.getElementById('myModal')
const myInput = document.getElementById('myInput')

myModal.addEventListener('shown.bs.modal', () => {
  myInput.focus()
})
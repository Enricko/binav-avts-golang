async function fetchUsers() {
    try {
        const response = await fetch('http://localhost:8080/user/data/autocomplete');
        const data = await response.json();
        const users = data.data;
        return users.map(user => `${user.id_user} | ${user.name}`); // Format as "idUser | name"
    } catch (error) {
        console.error('Error fetching users:', error);
        return [];
    }
}

$('#insertMapping').on('shown.bs.modal', async () =>{
    const userNames = await fetchUsers();

    const selectClient = new autoComplete({
        selector: "#selectClient",
        placeHolder: "Choose Client",
        data: {
            src: userNames,
            cache: true,
        },
        resultsList: {
            element: (list, data) => {
                if (!data.results.length) {
                    const message = document.createElement("div");
                    message.setAttribute("class", "no_result");
                    message.innerHTML = `<span>Found No Results for "${data.query}"</span>`;
                    list.prepend(message);
                }
            },
            noResults: true,
        },
        resultItem: {
            highlight: true
        },
        events: {
            input: {
                selection: (event) => {
                    const selection = event.detail.selection.value;
                    document.querySelector("#selectClient").value = selection;
                }
            }
        }
    });
});



function autoCompleteVesselSearch(data){
    const vesselSearch = new autoComplete({
        selector: "#vesselSearch",
        placeHolder: "Search...",
        data: {
            src: data,
            cache: true,
        },
        resultsList: {
            element: (list, data) => {
                if (!data.results.length) {
                    // Create "No Results" message element
                    const message = document.createElement("div");
                    // Add class to the created element
                    message.setAttribute("class", "no_result");
                    // Add message text content
                    message.innerHTML = `<span>Found No Results for "${data.query}"</span>`;
                    // Append message element to the results list
                    list.prepend(message);
                }
            },
            noResults: true,
        },
        resultItem: {
            highlight: true
        },
        events: {
            input: {
                selection: (event) => {
                    const selection = event.detail.selection.value;
                    vesselSearch.input.value = selection;
                }
            }
        }
    });
}

function updateAutoComplete(devices) {
    if (autoCompleteInstance) {
        autoCompleteInstance.data.src = devices;
        autoCompleteInstance.start();
    } else {
        autoCompleteVesselSearch(devices);
    }
}
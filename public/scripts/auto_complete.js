let autoCompleteInstance;
let vesselSearch;

function autoCompleteVesselSearch(data) {
  vesselSearch = new autoComplete({
    selector: "#vesselSearch",
    placeHolder: "Search...",
    data: {
      src: data,
      cache: false,
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
      highlight: true,
    },
    events: {
      input: {
        selection: (event) => {
          const selection = event.detail.selection.value;
          vesselSearch.input.value = selection;
        },
      },
    },
  });
}

function updateAutoComplete(devices) {
  if (vesselSearch) {
    vesselSearch.data.src = devices;
    vesselSearch.start();
  } else {
    autoCompleteVesselSearch(devices);
  }
}

// Global variables
let dataTableMap = {};

// Function to initialize or reload DataTable
function initializeDataTable(modalId, tableId, ajaxUrl, columns) {
  $(`#${modalId}`).on("shown.bs.modal", function () {
    let dataTable;

    if ($.fn.DataTable.isDataTable(`#${tableId}`)) {
      dataTable = $(`#${tableId}`).DataTable();
      dataTable.ajax.reload(null, false);
    } else {
      dataTable = $(`#${tableId}`).DataTable({
        ajax: ajaxUrl,
        processing: true,
        serverSide: true,
        searching: true,
        responsive: true,
        columnDefs: [
          { orderable: true, targets: 0 },
          { orderable: false, targets: "_all" },
        ],
        columns: columns,
      });
    }

    // Reload the table data every 60 seconds
    let reloadInterval = setInterval(() => {
      dataTable.ajax.reload(null, false);
    }, 60000);

    dataTableMap[tableId] = { dataTable, reloadInterval };
  });

  $(`#${modalId}`).on("hidden.bs.modal", function () {
    clearInterval(dataTableMap[tableId].reloadInterval);
  });
}

$(document).ready(function () {
  // Initialize DataTables
  



 
});

// Form submission handler
document
  .getElementById("submitClientButton")
  .addEventListener("click", function (event) {
    event.preventDefault();
    var formData = new FormData(document.getElementById("formInsertClient"));
    var json = JSON.stringify(Object.fromEntries(formData));

    fetch("/user/insert", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": formData.get("_csrf"),
      },
      body: json,
    })
      .then((response) => response.json())
      .then((data) => {
        if (response.ok) {
          dataTableMap["clientTableId"].dataTable.ajax.reload();
          $("#insertClient").modal("hide");
          document.getElementById("formInsertClient").reset();
          Swal.fire({
            title: "Success",
            text: data.message,
            icon: "success",
          });
        } else {
          throw new Error(data.message);
        }
      })
      .catch((error) => {
        Swal.fire({
          title: "Error",
          text: error.message,
          icon: "error",
        });
      });
  });

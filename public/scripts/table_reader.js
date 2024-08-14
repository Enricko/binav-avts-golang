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
  initializeDataTable("mappingTable", "mapping-table", "/mapping/data", [
    { data: "id_mapping", className: "text-center" },
    { data: "user.name", className: "text-center" },
    { data: "name", className: "text-center" },
    { data: "file", className: "text-center" },
    { data: "status", className: "text-center" },
    {
      data: "aksi",
      className: "text-center",
      width: "10%",
      render: function (data, type, row, meta) {
        return `
        <div class="container">
          <div class="row justify-content-md-center">
            <div class="col-auto mb-1">
              <button class="btn btn-warning btn-xs" onclick="updateForm(${row.ID})"><i class="fas fa-pencil"></i></button>
            </div>
            <div class="col-auto mb-1">
              <button class="btn btn-danger btn-xs" onclick="deleteData(${row.ID})"><i class="fas fa-trash-alt"></i></button>
            </div>
          </div>
        </div>
      `;
      },
    },
  ]);

  initializeDataTable("clientTable", "client-table", "/user/data", [
    { data: "id_user", className: "text-center" },
    { data: "name", className: "text-center" },
    {
      data: "send.email",
      className: "text-center",
      width: "20%",
      render: function (data, type, row, meta) {
        return `<button class="btn btn-primary btn-xs" onclick="updateForm(${row.ID})">Send Email</button>`;
      },
    },
    {
      data: "aksi",
      className: "text-center",
      width: "10%",
      render: function (data, type, row, meta) {
        return `
                  <div class="container">
                  <div class="row justify-content-md-center">
                  <div class="col-auto mb-1">
                  <button class="btn btn-warning btn-xs " id="edit_data" data-bs-toggle="modal" data-bs-target="#updateModal" onclick="updateForm(${row.ID})"><i class="fas fa-pencil"></i></button>
                  </div>
                  
                  <div class="col-auto mb-1">
                  <button class="btn btn-danger btn-xs " id="delete_data" onclick="deleteData(${row.ID})"><i class="fas fa-trash-alt"></i></button>
                  </div>
                  </div>
                  </div>
                  `;
      },
    },
  ]);

  initializeDataTable("vesselTable", "vessel-table", "/vessel/data", [
    { data: "call_sign", className: "text-center" },
    { data: "image", className: "text-center" },
    { data: "image_map", className: "text-center" },
    { data: "status", className: "text-center" },
    { data: "flag", className: "text-center" },
    { data: "kelas", className: "text-center" },
    { data: "builder", className: "text-center" },
    { data: "year_built", className: "text-center" },
    { data: "heading_direction", className: "text-center" },
    { data: "width_m", className: "text-center" },
    { data: "height_m", className: "text-center" },
    { data: "top_range", className: "text-center" },
    { data: "left_range", className: "text-center" },
    { data: "history_per_second", className: "text-center" },
    { data: "minimum_knot_per_liter_gasoline", className: "text-center" },
    { data: "maximum_knot_per_liter_gasoline", className: "text-center" },
    {
      data: "aksi",
      className: "text-center",
      width: "10%",
      render: function (data, type, row, meta) {
        return `
                  <div class="container">
                  <div class="row justify-content-md-center">
                  <div class="col-auto mb-1">
                  <button class="btn btn-warning btn-xs " id="edit_data" data-bs-toggle="modal" data-bs-target="#updateModal" onclick="updateForm(${row.ID})"><i class="fas fa-pencil"></i></button>
                  </div>
                  
                  <div class="col-auto mb-1">
                  <button class="btn btn-danger btn-xs " id="delete_data" onclick="deleteData(${row.ID})"><i class="fas fa-trash-alt"></i></button>
                  </div>
                  </div>
                  </div>
                  `;
      },
    },
  ]);
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

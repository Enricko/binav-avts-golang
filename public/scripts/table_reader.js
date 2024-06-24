let reloadInterval;
let itemTable;

//  ==== mappingTable ====
let dataTableMap = {};

let dataTableClient;

$('#mappingTable').on('shown.bs.modal', function () {
    let tableId = $(this).find('table').attr('id');
    let dataTable;

    if ($.fn.DataTable.isDataTable(`#${tableId}`)) {
        // If the table is already initialized, just reload the data
        dataTable = $(`#${tableId}`).DataTable();
        dataTable.ajax.reload(null, false);
    } else {
        // Otherwise, initialize the DataTable
        dataTable = $(`#${tableId}`).DataTable({
            ajax: '/mapping/data',
            processing: true,
            serverSide: true,
            searching: true, // Enable search feature
            responsive: true,
            columnDefs: [
                { "orderable": true, "targets": 0 },  // Enable sorting on the first column (First Name)
                { "orderable": false, "targets": '_all' }  // Disable sorting on all other columns
            ],
            columns: [
                { data: 'id_mapping', className: 'text-center' },
                { data: 'user.name', className: 'text-center' },
                { data: 'name', className: 'text-center' },
                { data: 'file', className: 'text-center' },
                { data: 'status', className: 'text-center' },
                { 
                    data: 'aksi',
                    className: 'text-center',
                    width: '10%',
                    render: function(data, type, row, meta){
                        return `
                        <div class="row">
                        <div class="col-auto mb-1">
                        <button class="btn btn-warning btn-xs " id="edit_data" data-bs-toggle="modal" data-bs-target="#updateModal" onclick="updateForm(${row.ID})"><i class="fas fa-pencil"></i></button>
                        </div>
                        
                        <div class="col-auto mb-1">
                        <button class="btn btn-danger btn-xs " id="delete_data" onclick="deleteData(${row.ID})"><i class="fas fa-trash-alt"></i></button>
                        </div>
                        </div>`;
                    }
                },
            ],
        });
    }

    // Reload the table data every 60 seconds
    let reloadInterval = setInterval(() => {
        dataTable.ajax.reload(null, false); // false to keep the current page
    }, 60000); // 60 seconds

    dataTableMap[tableId] = { dataTable, reloadInterval };
});

$('#mappingTable').on('hidden.bs.modal', function () {
    let tableId = $(this).find('table').attr('id');

    // Stop reloading the table data
    clearInterval(dataTableMap[tableId].reloadInterval);
});

 $('#clientTable').on('shown.bs.modal', function () {
    let tableId = $(this).find('table').attr('id');
    // let dataTable;

    if ($.fn.DataTable.isDataTable(`#${tableId}`)) {
        // If the table is already initialized, just reload the data
        dataTableClient = $(`#${tableId}`).DataTable();
        dataTableClient.ajax.reload(null, false);
    } else {
        // Otherwise, initialize the DataTable
        dataTableClient = $(`#${tableId}`).DataTable({
            ajax: '/user/data',
            processing: true,
            serverSide: true,
            searching: true, // Enable search feature
            responsive: true,
            columnDefs: [
                { "orderable": true, "targets": 0 },  // Enable sorting on the first column (First Name)
                { "orderable": false, "targets": '_all' }  // Disable sorting on all other columns
            ],
            columns: [
                { data: 'id_user', className: 'text-center' },
                { data: 'name', className: 'text-center' },
                { 
                    data: 'send.email',
                    className: 'text-center',
                    width: '20%',
                    render: function(data, type, row, meta){
                        return `
                        <button class="btn btn-primary btn-xs " id="edit_data" data-bs-toggle="modal" data-bs-target="#updateModal" onclick="updateForm(${row.ID})">Send Email</button>`;
                    }
                },
                { 
                    data: 'aksi',
                    className: 'text-center',
                    width: '10%',
                    render: function(data, type, row, meta){
                        return `
                        <div class="row">
                        <div class="col-auto mb-1">
                        <button class="btn btn-warning btn-xs " id="edit_data" data-bs-toggle="modal" data-bs-target="#updateModal" onclick="updateForm(${row.ID})"><i class="fas fa-pencil"></i></button>
                        </div>
                        
                        <div class="col-auto mb-1">
                        <button class="btn btn-danger btn-xs " id="delete_data" onclick="deleteData(${row.ID})"><i class="fas fa-trash-alt"></i></button>
                        </div>
                        </div>`;
                    }
                },
            ],
        });
    }

    // Reload the table data every 60 seconds
    let reloadInterval = setInterval(() => {
        dataTableClient.ajax.reload(null, false); // false to keep the current page
    }, 60000); // 60 seconds

    dataTableMap[tableId] = { dataTableClient, reloadInterval };
});

$('#clientTable').on('hidden.bs.modal', function () {
    let tableId = $(this).find('table').attr('id');

    // Stop reloading the table data
    clearInterval(dataTableMap[tableId].reloadInterval);
});

$('#vesselTable').on('shown.bs.modal', function () {
    let tableId = $(this).find('table').attr('id');
    let dataTable;

    if ($.fn.DataTable.isDataTable(`#${tableId}`)) {
        // If the table is already initialized, just reload the data
        dataTable = $(`#${tableId}`).DataTable();
        dataTable.ajax.reload(null, false);
    } else {
        // Otherwise, initialize the DataTable
        dataTable = $(`#${tableId}`).DataTable({
            ajax: '/vessel/data',
            processing: true,
            serverSide: true,
            searching: true, // Enable search feature
            responsive: true,
            columnDefs: [
                { "orderable": true, "targets": 0 },  // Enable sorting on the first column (First Name)
                { "orderable": false, "targets": '_all' }  // Disable sorting on all other columns
            ],
            columns: [
                { data: 'call_sign', className: 'text-center' },
                { data: 'status', className: 'text-center' },
                { data: 'flag', className: 'text-center' },
                { data: 'kelas', className: 'text-center' },
                { data: 'builder', className: 'text-center' },
                { data: 'year_built', className: 'text-center' },
                { data: 'heading_direction', className: 'text-center' },
                { data: 'size', className: 'text-center' },
                { data: 'xml_file', className: 'text-center' },
                { data: 'image', className: 'text-center' },
                { 
                    data: 'aksi',
                    className: 'text-center',
                    width: '20%',
                    render: function(data, type, row, meta){
                        return `
                        <div class="row">
                        <div class="col-auto mb-1">
                        <button class="btn btn-warning btn-xs " id="edit_data" data-bs-toggle="modal" data-bs-target="#updateModal" onclick="updateForm(${row.ID})"><i class="fas fa-pencil"></i></button>
                        </div>
                        
                        <div class="col-auto mb-1">
                        <button class="btn btn-danger btn-xs " id="delete_data" onclick="deleteData(${row.ID})"><i class="fas fa-trash-alt"></i></button>
                        </div>
                        </div>`;
                    }
                },
            ],
        });
    }

    // Reload the table data every 60 seconds
    let reloadInterval = setInterval(() => {
        dataTable.ajax.reload(null, false); // false to keep the current page
    }, 60000); // 60 seconds

    dataTableMap[tableId] = { dataTable, reloadInterval };
});

$('#vesselTable').on('hidden.bs.modal', function () {
    let tableId = $(this).find('table').attr('id');

    // Stop reloading the table data
    clearInterval(dataTableMap[tableId].reloadInterval);
});


// FUNCTION CREATE 
document.getElementById('submitClientButton').addEventListener('click', function(event) {
    var formData = FormData(document.getElementById("formInsertClient"));
    event.preventDefault();
    // var formData = new FormData(event.target);
    var json = JSON.stringify(Object.fromEntries(formData));
    fetch('/user/insert', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': formData.get('_csrf')
        },
        body: json
    }).then(response => {
        return response.json().then(data => {
            return { status: response.status, body: data };
        });
    }).then(data => {
        if (data.status < 300) {
            dataTableClient.ajax.reload();
            $('#insertClient').modal('hide');
            document.getElementById('formInsertClient').reset();
            Swal.fire({
                title: "Success",
                text: data.body.message,
                icon: "success"
            });
        }else{
            Swal.fire({
                title: "Error",
                text: data.body.message,
                icon: "error"
            });
        }
    }).catch(error => {
        Swal.fire({
            title: "Something went wrong",
            text: error.message,
            icon: "error"
        });
    });
});
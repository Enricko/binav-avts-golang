let reloadInterval;
let itemTable;

//  ==== mappingTable ====
let dataTableMap = {};

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
    let dataTable;

    if ($.fn.DataTable.isDataTable(`#${tableId}`)) {
        // If the table is already initialized, just reload the data
        dataTable = $(`#${tableId}`).DataTable();
        dataTable.ajax.reload(null, false);
    } else {
        // Otherwise, initialize the DataTable
        dataTable = $(`#${tableId}`).DataTable({
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
                    width: '10%',
                    render: function(data, type, row, meta){
                        return `
                        <button class="btn btn-warning btn-xs " id="edit_data" data-bs-toggle="modal" data-bs-target="#updateModal" onclick="updateForm(${row.ID})">Send Email</button>`;
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
        dataTable.ajax.reload(null, false); // false to keep the current page
    }, 60000); // 60 seconds

    dataTableMap[tableId] = { dataTable, reloadInterval };
});

$('#clientTable').on('hidden.bs.modal', function () {
    let tableId = $(this).find('table').attr('id');

    // Stop reloading the table data
    clearInterval(dataTableMap[tableId].reloadInterval);
});
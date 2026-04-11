const pageContent = document.querySelector(".page-content");
const logoutBtn = document.querySelector(".logout-btn");
const pageTitle = document.querySelector(".page-title b");
const menuItems = document.querySelectorAll(".menu-item");
const profileName = document.querySelector(".top-panel .profile-name");

let currentPageLoaded = "home";
let currentUserRole = "";

async function getCurrentUser() {
    const response = await fetch("/api/profile/me");

    if (!response.ok) {
        throw new Error("Не удалось получить данные текущего пользователя");
    }

    const result = await response.json();
    return result.data;
}

function applyRoleVisibility(root, currentUserRole) {
    const roleElements = root.querySelectorAll("[data-role]");

    for (const element of roleElements) {
        const allowedRoles = element.dataset.role;

        if (!allowedRoles) {
            continue;
        }

        const roles = allowedRoles.split(" ");
        const isAllowed = roles.includes(currentUserRole);

        element.hidden = !isAllowed;
    }
}

async function loadPage(pageName) {
    try {
        const response = await fetch(`../static/partials/${pageName}-content.html`);

        if (!response.ok) {
            throw new Error("Страница не найдена");
        }

        const html = await response.text();
        pageContent.innerHTML = html;

        if (currentUserRole) {
            applyRoleVisibility(pageContent, currentUserRole);
        }

        currentPageLoaded = pageName;
        await initialization(pageName);

        console.log(`${pageName} page initialized`);
    } catch (error) {
        pageContent.innerHTML = "<h2>Ошибка загрузки страницы</h2>";
        console.error(error);
    }
}

async function initApp() {
    const activeMenuItem = document.querySelector(".menu-item.active");

    if (!activeMenuItem) {
        return;
    }

    const page = activeMenuItem.dataset.page;
    pageTitle.textContent = activeMenuItem.textContent.trim();

    const profile = await getCurrentUser();
    currentUserRole = profile.user.role;

    const menu = document.querySelector(".menu");
    applyRoleVisibility(menu, currentUserRole);

    if (profileName) {
        profileName.textContent = profile.user.full_name;
    }

    await loadPage(page);
}

menuItems.forEach(item => {
    item.addEventListener("click", async event => {
        event.preventDefault();

        const currentItem = event.currentTarget;
        const page = currentItem.dataset.page;

        pageTitle.textContent = currentItem.textContent.trim();
        await loadPage(page);

        document.querySelector(".menu-item.active")?.classList.remove("active");
        currentItem.classList.add("active");
    });
});

if (logoutBtn) {
    logoutBtn.addEventListener("click", async event => {
        event.preventDefault();

        try {
            await logout();
        } catch (error) {
            console.error(error);
            alert("Не удалось выйти из аккаунта");
        }
    });
}

async function initialization(currentPageLoaded) {
    if (currentPageLoaded === "home") {
        homePageHandler();
    } else if (currentPageLoaded === "orders") {
        await ordersPageHandler();
    } else if (currentPageLoaded === "cars") {
        await carsPageHandler();
    } else if (currentPageLoaded === "storage"){
        await storagePageHandler();
    }
}

function homePageHandler() {
}

async function ordersPageHandler() {
    const table = pageContent.querySelector(".orders-table");
    const tableBody = pageContent.querySelector("tbody");

    if (!table || !tableBody) {
        return;
    }

    async function getOrders() {
        const response = await fetch("/api/orders");

        if (!response.ok) {
            throw new Error("Не удалось получить заказы");
        }

        const result = await response.json();
        return result.data;
    }

    function getEmployeesText(employees) {
        if (!employees || employees.length === 0) {
            return "—";
        }

        return employees.map(employee => employee.full_name).join(", ");
    }

    function getVisibleColumnsCount(table) {
        return table.querySelectorAll("thead th:not([hidden])").length;
    }

    function renderTableMessage(message) {
        const visibleColumnsCount = getVisibleColumnsCount(table);

        tableBody.innerHTML = `
            <tr>
                <td colspan="${visibleColumnsCount}">${message}</td>
            </tr>
        `;
    }

    function renderOrders(orders, tableBody) {
        tableBody.innerHTML = orders.map(order => `
            <tr>
                <td>${order.order_number ?? "—"}</td>
                <td data-role="master admin">${order.owner_full_name ?? "—"}</td>
                <td>${order.car_plate_number ?? "—"}</td>
                <td>${order.service_name ?? "—"}</td>
                <td>${order.owner_phone ?? "—"}</td>
                <td data-role="admin">${getEmployeesText(order.employees)}</td>
                <td>${order.ready_date ?? "—"}</td>
            </tr>
        `).join("");

        applyRoleVisibility(tableBody, currentUserRole);
    }

    try {
        const orders = await getOrders();
        console.log(orders);

        if (!orders.length) {
            renderTableMessage("Заказов пока нет");
            return;
        }

        renderOrders(orders, tableBody);
    } catch (error) {
        console.error(error);
        renderTableMessage("Не удалось загрузить заказы");
    }
}

async function carsPageHandler () {
    const table = pageContent.querySelector(".cars-table");

    if (!table) {
        return;
    }

    const tableBody = table.querySelector("tbody");

    if (!tableBody){
        return;
    }
    
    async function getCars() {
        const response = await fetch("/api/cars");

        if (!response.ok){
            throw new Error("Не удалось получить данные о машинах и их владельцах");
        }

        const result = await response.json();
        return result.data;
    }

    function getVisibleColumnsCount(table) {
        return table.querySelectorAll("thead th:not([hidden])").length;
    }

    function renderTableMessage(message) {
        const visibleColumnsCount = getVisibleColumnsCount(table);

        tableBody.innerHTML = `
            <tr>
                <td colspan="${visibleColumnsCount}">${message}</td>
            </tr>
        `;
    }

    function renderCars(cars, tableBody) {
        tableBody.innerHTML = cars.map (car => 
            `<tr>
                <td>${car.owner_full_name ?? "—"}</td>
                <td>${car.brand ?? "—"}</td>
                <td>${car.plate_number ?? "—"}</td>
                <td>${car.manufacture_year ?? "—"}</td>
            </tr>`
            ).join("");            
    }

    try {
        const cars = await getCars();
        console.log(cars);

        if (!cars.length) {
            renderTableMessage("Данных о машинах пока нет");
            return;
        }

        renderCars(cars, tableBody);
    } catch (error) {
        console.error(error);
        renderTableMessage("Не удалось загрузить данные о машинах и владельцах");
    }

}

async function storagePageHandler() {
    const table = pageContent.querySelector(".storage-table");

    if (!table) {
        return;
    }

    const tableBody = table.querySelector("tbody");
    const tableTemplate = pageContent.querySelector("#storage-row-template");

    const popup = pageContent.querySelector(".storage-popup");
    const popupForm = pageContent.querySelector(".storage-order-form");
    const popupCancelBtn = pageContent.querySelector(".storage-popup-cancel");
    const popupBackdrop = pageContent.querySelector(".storage-popup-backdrop");

    if (!tableBody || !tableTemplate || !popup || !popupForm || !popupCancelBtn || !popupBackdrop) {
        return;
    }

    let activeStorageRow = null;

        function getVisibleColumnsCount(table) {
        return table.querySelectorAll("thead th:not([hidden])").length;
    }

    function renderTableMessage(message) {
        const visibleColumnsCount = getVisibleColumnsCount(table);

        tableBody.innerHTML = `
            <tr>
                <td colspan="${visibleColumnsCount}">${message}</td>
            </tr>
        `;
    }

    try{
        async function getRepairParts() {
        const response = await fetch("/api/repair-parts");

        if (!response.ok){
            throw new Error("Не удалось получить список запчастей.");
        }

        const result = await response.json();
        return result.data;
    }
    const repairParts = await getRepairParts();

    async function getMaintenanceParts() {
        const response = await fetch("/api/maintenance-parts");

        if (!response.ok){
            throw new Error("Не удалось получить список расходников.")
        }

        const result = await response.json();
        return result.data;
    }
    const maintenanceParts = await getMaintenanceParts();

    const normalizedRepairParts = repairParts.map(part => ({
        id: part.repair_part_id,
        type: "repair",
        name: part.name,
        quantity: part.quantity,
        deliveryDate: part.delivery_date,
    }));

    const normalizedMaintenanceParts = maintenanceParts.map(part => ({
        id: part.maintenance_part_id,
        type: "maintenance",
        name: part.name,
        quantity: part.quantity,
        deliveryDate: part.delivery_date,
    }));

    const storageItems = [...normalizedRepairParts, ...normalizedMaintenanceParts];

    if (storageItems.length === 0){
        renderTableMessage("Склад пуст или товары не найдены");
        return;
    }

    tableBody.innerHTML = ""; 
    renderStorageItems(storageItems, tableBody);

    function renderStorageItems (storageItems, tableBody) {
        storageItems.forEach(item => {
            const clone = tableTemplate.content.cloneNode(true);

            const row = clone.querySelector(".storage-row");

            row.dataset.id = item.id;
            row.dataset.type = item.type;

            row.querySelector(".part-name").textContent = item.name ?? "—";
            row.querySelector(".part-quantity").textContent = item.quantity ?? "—";
            row.querySelector(".part-delivery-date").textContent = item.deliveryDate ?? "—";

            tableBody.append(clone);
        });
    }} catch (error) {
        console.error(error);
        renderTableMessage("Не удалось загрузить содержимое склада");
    }

    function openStoragePopup(row) {
        activeStorageRow = row;
        popup.hidden = false;
    }

    function closeStoragePopup() {
        activeStorageRow = null;
        popupForm.reset();
        popup.hidden = true;
    }

    tableBody.addEventListener("click", handleStorageTableClick);
    popupCancelBtn.addEventListener("click", closeStoragePopup);
    popupBackdrop.addEventListener("click", closeStoragePopup);
    popupForm.addEventListener("submit", handleStorageFormSubmit);

    function handleStorageTableClick(event) {
        const button = event.target.closest("button");

        if (!button) {
            return;
        }

        const row = button.closest(".storage-row");

        if (!row) {
            return;
        }

        const action = button.dataset.action;
        const itemId = row.dataset.id;
        const itemType = row.dataset.type;

        console.log(action, itemId, itemType);

        openStoragePopup(row);
    }

    async function handleStorageFormSubmit(event) {
        event.preventDefault();

        if (!activeStorageRow) {
            return;
        }

        const formData = new FormData(popupForm);
        const addedQuantity = Number(formData.get("quantity"));
        const deliveryDate = formData.get("delivery_date");

        if (!addedQuantity || addedQuantity < 1) {
            return;
        }

        const itemId = activeStorageRow.dataset.id;
        const itemType = activeStorageRow.dataset.type;

        let url = "";

        if (itemType === "repair") {
            url = `/api/repair-parts/${itemId}/restock`;
        } else if (itemType === "maintenance") {
            url = `/api/maintenance-parts/${itemId}/restock`;
        } else {
            console.error("Неизвестный тип позиции склада:", itemType);
            return;
        }

        try {
            const response = await fetch(url, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    quantity: addedQuantity,
                    delivery_date: deliveryDate,
                }),
            });

            if (!response.ok) {
                throw new Error("Не удалось сохранить изменения на сервере");
            }

            const result = await response.json();
            const updatedItem = result.data;

            const quantityCell = activeStorageRow.querySelector(".part-quantity");
            const dateCell = activeStorageRow.querySelector(".part-delivery-date");

            quantityCell.textContent = updatedItem.quantity ?? "—";
            dateCell.textContent = updatedItem.delivery_date ?? "—";

            closeStoragePopup();
        } catch (error) {
            console.error(error);
            alert("Не удалось сохранить изменения");
        }
    }
}

async function checkAuth() {
    try {
        const response = await fetch("/api/profile/me");

        if (!response.ok) {
            window.location.href = "/auth";
            return false;
        }

        return true;
    } catch (error) {
        console.error(error);
        window.location.href = "/auth";
        return false;
    }
}

async function logout() {
    const response = await fetch("/api/auth/logout", {
        method: "POST",
        credentials: "same-origin",
    });

    if (!response.ok) {
        throw new Error("Не удалось выполнить выход");
    }

    window.location.href = "/auth";
}

async function startApp() {
    const isAuthorized = await checkAuth();

    if (!isAuthorized) {
        return;
    }

    await initApp();
}

startApp();
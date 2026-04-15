const pageContent = document.querySelector(".page-content");
const logoutBtn = document.querySelector(".logout-btn");
const pageTitle = document.querySelector(".page-title b");
const menuItems = document.querySelectorAll(".menu-item");
const profileName = document.querySelector(".top-panel .profile-name");

let currentPageLoaded = "home";
let currentUserProfile = null;
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

    currentUserProfile = profile;

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
    } else if (currentPageLoaded === "storage") {
        await storagePageHandler();
    } else if (currentPageLoaded === "profile") {
        await profilePageHandler();
    }
}

async function homePageHandler() {
    const cards = pageContent.querySelectorAll(".mid-card");

    if (cards.length < 4) {
        return;
    }

    const ordersEl = cards[0].querySelector(".card-stat-text b");
    const ownersEl = cards[1].querySelector(".card-stat-text b");
    const carsEl = cards[2].querySelector(".card-stat-text b");
    const storageEl = cards[3].querySelector(".card-stat-text b");

    if (!ordersEl || !ownersEl || !carsEl || !storageEl) {
        return;
    }

    try {
        const [
            ordersRes,
            ownersRes,
            carsRes,
            repairPartsRes,
            maintenancePartsRes
        ] = await Promise.all([
            fetch("/api/orders"),
            fetch("/api/owners"),
            fetch("/api/cars"),
            fetch("/api/repair-parts"),
            fetch("/api/maintenance-parts")
        ]);

        if (
            !ordersRes.ok ||
            !ownersRes.ok ||
            !carsRes.ok ||
            !repairPartsRes.ok ||
            !maintenancePartsRes.ok
        ) {
            throw new Error("Ошибка загрузки статистики");
        }

        const ordersResult = await ordersRes.json();
        const ownersResult = await ownersRes.json();
        const carsResult = await carsRes.json();
        const repairPartsResult = await repairPartsRes.json();
        const maintenancePartsResult = await maintenancePartsRes.json();

        const orders = ordersResult.data ?? [];
        const owners = ownersResult.data ?? [];
        const cars = carsResult.data ?? [];
        const repairParts = repairPartsResult.data ?? [];
        const maintenanceParts = maintenancePartsResult.data ?? [];

        const ordersCount = orders.length;
        const ownersCount = owners.length;
        const carsCount = cars.length;
        const storageCount = repairParts.length + maintenanceParts.length;

        ordersEl.textContent = ordersCount;
        ownersEl.textContent = ownersCount;
        carsEl.textContent = carsCount;
        storageEl.textContent = storageCount;
    } catch (error) {
        console.error("Ошибка загрузки главной страницы:", error);

        ordersEl.textContent = "—";
        ownersEl.textContent = "—";
        carsEl.textContent = "—";
        storageEl.textContent = "—";
    }
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
                <td data-role="master">${order.owner_full_name ?? "—"}</td>
                <td>${order.car_plate_number ?? "—"}</td>
                <td>${order.service_name ?? "—"}</td>
                <td>${order.owner_phone ?? "—"}</td>
                <td>${order.ready_date ?? "—"}</td>
            </tr>
        `).join("");

        applyRoleVisibility(tableBody, currentUserRole);
    }

    async function refreshOrders() {
        const orders = await getOrders();
        console.log(orders);

        if (!orders.length) {
            renderTableMessage("Заказов пока нет");
            return;
        }
        renderOrders(orders, tableBody);
    }
        
    await refreshOrders();

    if (currentUserRole !== "master") {
        return;
    }

    const createOrderBtn = pageContent.querySelector(".create-order");
    const ordersPopup = pageContent.querySelector(".popup-create-order");
    const orderForm = pageContent.querySelector(".create-order-form");

    if (!createOrderBtn || !ordersPopup || !orderForm) {
        return;
    }

    const formMaster =  orderForm.querySelector("#order-master");
    const formOrderCar = orderForm.querySelector("#order-car");
    const formOrderOwner = orderForm.querySelector("#order-owner");
    const formOrderService = orderForm.querySelector("#order-service");
    const formOrderDate = orderForm.querySelector("#order-ready-date");

    const formCloseBtn = orderForm.querySelector(".cancel-create-order-btn");

    const masterUser = currentUserProfile.user;
    const masterEmployee = currentUserProfile.employee;

    const masterName = masterUser.full_name;
    const masterId = masterEmployee.employee_id;

    async function getFormData() {
        const response = await fetch("/api/orders/form-data");

        if(!response.ok){
            return;
        }

        const result = await response.json();
        return result.data;
    }

    const orderFormData = await getFormData();

    function fillCarSelect(){
        formOrderCar.innerHTML = "";
    
        const defaultOption = document.createElement("option");
        defaultOption.value = "";
        defaultOption.textContent = "Выберите авто";
        formOrderCar.append(defaultOption);

        for(const car of orderFormData.cars){
            const option = document.createElement("option");
            option.value = car.car_id;
            option.textContent = car.label;
            formOrderCar.append(option);
        }
    }

    function fillServiceSelect(){
        formOrderService.innerHTML = "";

        const defaultOption = document.createElement("option");
        defaultOption.value = "";
        defaultOption.textContent = "Выберите услугу";
        formOrderService.append(defaultOption);

        for(const service of orderFormData.services){
            const option = document.createElement("option");
            option.value = service.service_id;
            option.textContent = service.name;
            formOrderService.append(option);
        }
    }

    formOrderCar.addEventListener("change", () => {
        const selectedCarId = +formOrderCar.value;
        const selectedCar = orderFormData.cars.find(car => car.car_id === selectedCarId);

        if (!selectedCar){
            formOrderOwner.value = "";
            return;
        }
        
        formOrderOwner.value = selectedCar.owner_full_name;
    });

    createOrderBtn.addEventListener("click", event => {
        event.preventDefault();

        orderForm.reset();
        formMaster.value = masterName;

        formOrderOwner.value = "";
        fillCarSelect();
        fillServiceSelect();

        ordersPopup.classList.remove("hidden");
    });

    formCloseBtn.addEventListener("click", () => {
        ordersPopup.classList.add("hidden");
    });

    orderForm.addEventListener("submit", async (event) => {
        event.preventDefault();

        const formData = {
            car_id: +formOrderCar.value,
            service_id: +formOrderService.value,
            ready_date: formOrderDate.value
        };

        try {
            const createResponse = await fetch("/api/orders", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(formData)
            });

            if (!createResponse.ok) {
                const errorData = await createResponse.json().catch(() => null);
                console.error("Ошибка создания заказа:", errorData);
                throw new Error("Ошибка при создании заказа");
            }

            const createdResult = await createResponse.json();
            const createdOrder = createdResult.data;

            const assignResponse = await fetch("/api/orders/assign-employees", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    order_number: createdOrder.order_number,
                    employee_id: masterId
                })
            });

            if (!assignResponse.ok) {
                const errorData = await assignResponse.json().catch(() => null);
                console.error("Ошибка привязки мастера:", errorData);
                throw new Error("Заказ создан, но не удалось привязать мастера");
            }

            ordersPopup.classList.add("hidden");
            await refreshOrders();
            alert("Заказ успешно создан!");
        } catch (error) {
            console.error(error);
            alert(error.message || "Не удалось сохранить заказ");
        }
    });

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

async function profilePageHandler() {
    const profileName = pageContent.querySelector(".profile-name");
    const profileRole = pageContent.querySelector(".user-role");

    const profileEmail = pageContent.querySelector(".user-email");
    const profilePhone = pageContent.querySelector(".user-phone");

    const userCars = pageContent.querySelector(".user-added-cars");
    const carTemplate = pageContent.querySelector("#car-card-template");

    const profile = await getCurrentUser();

    profileName.textContent = profile.user.full_name;

    if (profile.user.role === "master"){
        profileRole.textContent = "Мастер";
    }

    profileEmail.textContent = profile.user.email;
    profilePhone.textContent = profile.user.phone;

    userCars.addEventListener("click", async (event) => {
        const deleteButton = event.target.closest('[data-action="delete"]');
        if (!deleteButton){
            return;
        }
        const carToDelete = deleteButton.closest(".car-card");
        if(!carToDelete){
            return;
        }

        const carId = carToDelete.dataset.id;
        if(!carId){
            return;
        }
        
       const response = await fetch(`/api/cars/${carId}`,{
        method: "DELETE"
       });
       if (!response.ok){
        throw new Error("Не удалось удалить машину");
       }
       await carsListRender()
    });

    const createCar = pageContent.querySelector(".add-car");
    const popup = pageContent.querySelector(".pop-up-create-car");
    const closePopup = pageContent.querySelector("#close-popup");
    const addCarForm = pageContent.querySelector("#add-car-form");

    function openCarPopup() {
        popup.classList.remove("hidden");
    }

    function closeCarPopup() {
        popup.classList.add("hidden");
        addCarForm.reset();
    }

    if (!profile.owner){
        return;
    }

    const ownerId = profile.owner.owner_id;

    createCar.addEventListener("click", openCarPopup);
    closePopup.addEventListener("click", closeCarPopup);

     addCarForm.addEventListener("submit", async (event) => {
        event.preventDefault();

        if(!ownerId){
            return;
        }

        const formData = new FormData(addCarForm);
        const carBrand = formData.get("brand").trim();
        const carPlateNumber = formData.get("plate_number").trim();
        const carManufactureYear = +formData.get("manufacture_year");
        const carColor = formData.get("color").trim();

        const carData = {
            brand: carBrand,
            plate_number: carPlateNumber,
            manufacture_year: carManufactureYear,
            color: carColor,
            owner_id: ownerId
        }

        const response = await fetch("/api/cars",{
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(carData)
        });

        if (!response.ok) {
            throw new Error("Не удалось создать запись о машине");
        }
        closeCarPopup();
        await carsListRender();
    });

    async function getOwnerCars(ownerId) {
        const response = await fetch("/api/cars");
        if(!response.ok){
            throw new Error("Не удалось получить список машин");
        }

        const result = await response.json();
        const cars = result.data;

        const ownerCars = [];

        for (const car of cars){
            if (car.owner_id === ownerId){
                ownerCars.push(car);
            }            
        }
        return ownerCars;
    }

    async function carsListRender() {
        const currentOwnerCars = await getOwnerCars(ownerId);

        if(!userCars || !carTemplate){
            return;
        }

        userCars.innerHTML = "";

        if(currentOwnerCars.length === 0){
            userCars.textContent = "На данный момент не добавлено ни одной машины";
            return;
        }

        currentOwnerCars.forEach(car => {
            const carCard = carTemplate.content.cloneNode(true);
            
            const carCardClone = carCard.querySelector(".car-card");
            carCardClone.dataset.id = car.car_id;

            const carBrand = carCard.querySelector(".car-brand");
            const carPlate = carCard.querySelector(".car-plate");
            const carYear = carCard.querySelector(".car-year");
            const carColor = carCard.querySelector(".car-color");

            carBrand.textContent = car.brand ?? "—";
            carPlate.textContent = car.plate_number ?? "—";
            carYear.textContent = car.manufacture_year ?? "—";
            carColor.textContent = car.color ?? "—";

            userCars.append(carCard);   
        });
    }
    await carsListRender();
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
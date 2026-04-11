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
    const tableBody = table.querySelector("tbody");

    if (!table || !tableBody) {
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
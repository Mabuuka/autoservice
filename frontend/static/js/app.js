const pageContent = document.querySelector(".page-content");
const pageTitle = document.querySelector(".page-title b");
const menuItems = document.querySelectorAll(".menu-item");

let currentPageLoaded ="home";

async function loadPage(pageName){
    try{
        const response = await fetch(`../static/partials/${pageName}-content.html`);
        
        if (!response.ok) {
            throw new Error("Страница не найдена");
        }

        const html = await response.text();
        pageContent.innerHTML = html;


        currentPageLoaded = pageName;
        initialization(pageName);
        console.log(`${pageName} page initialized`);
    } catch (error){
        pageContent.innerHTML = '<h2>Ошибка загрузки страницы</h2>';
        console.error(error);
    }
}

function initApp(){
    const activeMenuItem = document.querySelector(".menu-item.active");

    if (!activeMenuItem) return;

    const page = activeMenuItem.dataset.page;
    pageTitle.textContent = activeMenuItem.textContent.trim();
    loadPage(page);
}

menuItems.forEach(item => {
    item.addEventListener("click", (event) =>{
        event.preventDefault();

        const currentItem = event.currentTarget;
        const page = currentItem.dataset.page;
        pageTitle.textContent = currentItem.textContent.trim();
        loadPage(page);

        document.querySelector('.menu-item.active')?.classList.remove('active');
        currentItem.classList.add('active');
    });
});

function initialization(currentPageLoaded){
    if (currentPageLoaded === "home"){
        homePageHandler();
    } else if (currentPageLoaded === "orders") {
        ordersPageHandler();
    }
}

function homePageHandler() {
}

async function ordersPageHandler() {
    const tableBody = pageContent.querySelector("tbody");

    async function getOrders() {
    const response = await fetch("/api/orders");

    if (!response.ok) {
        throw new Error("Не удалось получить заказы");
    }

    const result = await response.json();
    return result.data;
    }

    const orders = await getOrders();
    console.log(orders);

    function renderOrders(orders, tableBody) {
    tableBody.innerHTML = orders.map(order => `
        <tr>
            <td>${order.order_number}</td>
            <td>${order.car_plate_number}</td>
            <td>${order.owner_full_name}</td>
            <td>${order.owner_phone}</td>
            <td>${order.service_name}</td>
            <td>${order.ready_date}</td>
        </tr>
    `).join("");
    }
    renderOrders(orders, tableBody);
}

async function checkAuth() {
const response = await fetch("/api/profile/me");

if (!response.ok) {
    window.location.href = "/auth";
    return false;
}

return true;
}

async function startApp() {
    const isAuthorized = await checkAuth();

    if (!isAuthorized) return;

    initApp();
}

startApp();

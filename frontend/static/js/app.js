const pageContent = document.querySelector(".page-content");
const pageTitle = document.querySelector(".page-title b");
const menuItems = document.querySelectorAll(".menu-item");

async function loadPage(pageName){
    try{
        const response = await fetch(`../static/partials/${pageName}-content.html`);
        
        if (!response.ok) {
            throw new Error("Страница не найдена");
        }

        const html = await response.text();
        pageContent.innerHTML = html;

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

initApp();
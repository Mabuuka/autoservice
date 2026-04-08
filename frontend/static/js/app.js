const pageContent = document.querySelector(".page-content");
const pageTitle = document.querySelector(".page-title b");
const menuItem = document.querySelectorAll(".menu-item");

async function loadPage(pageName){
    try{

        const responce = await fetch(`../static/partials/${pageName}-content.html`);
        const html = await responce.text();
        pageContent.innerHTML = html;

    } catch (error){
        pageContent.innerHTML = '<h2>Ошибка загрузки страницы</h2>';
        console.error(error);
    }
}


menuItem.forEach( item => {
    item.addEventListener("click", (event) =>{
        event.preventDefault();

        const page = item.getAttribute('data-page');
        pageTitle.textContent = item.textContent;
        loadPage(page);

        document.querySelector('.menu-item.active')?.classList.remove('active');
        item.classList.add('active');
    });
});
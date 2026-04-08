const emailInput = document.querySelector("#email_login");
const passwordInput = document.querySelector("#password_login");
const loginButton = document.querySelector("#login_button");

const regNameInput = document.querySelector("#name-reg");
const regEmailInput = document.querySelector("#email-reg");
const regPhoneInput = document.querySelector("#phone-reg");
const regLicenseInput = document.querySelector("#license-reg");
const regPasswordInput = document.querySelector("#pass-reg");
const regPasswordRepeatInput = document.querySelector("#pass-sec-reg");
const registrationButton = document.querySelector("#registration_button");

async function getResponseData(response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

function getErrorMessage(data, fallbackText) {
    if (!data) return fallbackText;

    return (
        data.message ||
        data.error?.message ||
        data.error ||
        fallbackText
    );
}

if (loginButton) {
    loginButton.addEventListener("click", async () => {
        const email = emailInput.value.trim();
        const password = passwordInput.value.trim();

        if (!email || !password) {
            alert("Введите почту и пароль.");
            return;
        }

        try {
            const response = await fetch("/api/auth/login", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    email,
                    password
                })
            });

            const data = await getResponseData(response);

            if (!response.ok) {
                alert(getErrorMessage(data, "Ошибка входа."));
                return;
            }

            window.location.href = "/";
        } catch (error) {
            console.error("Login error:", error);
            alert("Не удалось выполнить вход. Проверь, запущен ли сервер.");
        }
    });
}

if (registrationButton) {
    registrationButton.addEventListener("click", async () => {
        const fullName = regNameInput.value.trim();
        const email = regEmailInput.value.trim();
        const phone = regPhoneInput.value.trim();
        const driverLicense = regLicenseInput.value.trim();
        const password = regPasswordInput.value.trim();
        const repeatedPassword = regPasswordRepeatInput.value.trim();

        if (!fullName || !email || !phone || !driverLicense || !password || !repeatedPassword) {
            alert("Заполни все поля регистрации.");
            return;
        }

        if (password !== repeatedPassword) {
            alert("Пароли не совпадают.");
            return;
        }

        try {
            const response = await fetch("/api/auth/register", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    full_name: fullName,
                    email,
                    phone,
                    driver_license: driverLicense,
                    password
                })
            });

            const data = await getResponseData(response);

            if (!response.ok) {
                alert(getErrorMessage(data, "Ошибка регистрации."));
                return;
            }

            window.location.href = "/";
        } catch (error) {
            console.error("Registration error:", error);
            alert("Не удалось выполнить регистрацию. Проверь, запущен ли сервер.");
        }
    });
}
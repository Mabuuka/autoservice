BEGIN;

SET search_path TO automaster;

TRUNCATE TABLE
    order_maintenance_parts,
    order_repair_parts,
    order_employees,
    orders,
    employee_salaries,
    maintenance_parts,
    repair_parts,
    cars,
    employees,
    services,
    owners
RESTART IDENTITY CASCADE;

-- Владельцы
INSERT INTO owners (full_name, phone, driver_license, is_regular) VALUES
    ('Илья Соколов', '+7-900-111-22-33', '77 01 123456', true),
    ('Анна Морозова', '+7-900-222-33-44', '77 02 234567', false),
    ('Дмитрий Волков', '+7-900-333-44-55', '77 03 345678', true),
    ('Елена Кузнецова', '+7-900-444-55-66', '77 04 456789', false);

-- Автомобили
INSERT INTO cars (owner_id, brand, plate_number, manufacture_year, color) VALUES
    (1, 'Toyota Camry', 'А123ВС777', 2018, 'черный'),
    (2, 'Kia Rio', 'В456КХ777', 2020, 'белый'),
    (3, 'LADA Vesta', 'С789МН777', 2022, 'серый'),
    (4, 'Hyundai Solaris', 'Е321ОР777', 2019, 'синий'),
    (1, 'Skoda Octavia', 'К654ТТ777', 2017, 'серебристый');

-- Услуги
INSERT INTO services (name, description, price_rub, regular_discount_percent) VALUES
    ('Диагностика двигателя', 'Компьютерная диагностика и проверка основных узлов двигателя.', 2500.00, 10.00),
    ('Замена масла', 'Плановое техническое обслуживание с заменой масла и фильтров.', 1800.00, 15.00),
    ('Замена тормозных колодок', 'Снятие изношенных колодок и установка нового комплекта.', 3200.00, 5.00),
    ('Развал-схождение', 'Настройка углов установки колес на стенде.', 2200.00, 0.00),
    ('Кузовная полировка', 'Восстановительная полировка лакокрасочного покрытия.', 6000.00, 20.00);

-- Сотрудники
INSERT INTO employees (personnel_number, specialty, phone, full_name) VALUES
    (101, 'Автомеханик', '+7-901-100-10-10', 'Алексей Петров'),
    (102, 'Диагност', '+7-901-200-20-20', 'Марина Смирнова'),
    (103, 'Мастер-приемщик', '+7-901-300-30-30', 'Олег Васильев'),
    (104, 'Слесарь по ТО', '+7-901-400-40-40', 'Ирина Орлова');

-- Зарплата сотрудников
INSERT INTO employee_salaries (employee_id, participation_coeff, full_shift_salary_rub) VALUES
    (1, 1.10, 4500.00),
    (2, 1.20, 4800.00),
    (3, 1.00, 4200.00),
    (4, 1.05, 4300.00);

-- Детали для ремонта
INSERT INTO repair_parts (name, quantity, delivery_date) VALUES
    ('Тормозные колодки передние', 12, '2026-03-20'),
    ('Амортизатор передний', 6, '2026-03-18'),
    ('Ремень генератора', 10, '2026-03-25'),
    ('Свечи зажигания, комплект', 8, '2026-03-28');

-- Детали для ТО
INSERT INTO maintenance_parts (name, quantity, delivery_date) VALUES
    ('Масло моторное 5W-30', 40, '2026-03-15'),
    ('Масляный фильтр', 20, '2026-03-16'),
    ('Воздушный фильтр', 15, '2026-03-18'),
    ('Салонный фильтр', 15, '2026-03-19');

-- Заказы
INSERT INTO orders (car_id, service_id, ready_date) VALUES
    (1, 1, '2026-04-05'),
    (2, 2, '2026-04-06'),
    (3, 3, '2026-04-07'),
    (5, 4, '2026-04-09'),
    (4, 5, '2026-04-12');

-- Сотрудники в заказах
INSERT INTO order_employees (order_number, employee_id) VALUES
    (1, 2),
    (1, 3),
    (2, 1),
    (2, 4),
    (3, 1),
    (3, 3),
    (4, 2),
    (5, 1),
    (5, 2);

-- Детали для ремонта в заказах
INSERT INTO order_repair_parts (order_number, repair_part_id, quantity_used) VALUES
    (3, 1, 1),
    (3, 4, 1);

-- Детали для ТО в заказах
INSERT INTO order_maintenance_parts (order_number, maintenance_part_id, quantity_used) VALUES
    (2, 1, 4),
    (2, 2, 1),
    (2, 3, 1),
    (2, 4, 1);

COMMIT;

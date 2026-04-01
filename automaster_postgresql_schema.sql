create schema if not exists automaster;
set search_path to automaster;

-- =========================
-- 1. Владельцы
-- =========================
create table owners (
    owner_id bigint generated always as identity primary key,
    full_name varchar(200) not null unique,
    phone varchar(20) not null unique,
    driver_license varchar(50) not null,
    is_regular boolean not null default false
);

-- =========================
-- 2. Автомобили
-- =========================
create table cars (
    car_id bigint generated always as identity primary key,
    owner_id bigint not null references owners(owner_id)
        on update cascade
        on delete restrict,
    brand varchar(100) not null,
    plate_number varchar(20) not null unique,
    manufacture_year integer not null
        check (
            manufacture_year between 1900
            and extract(year from current_date)::int + 1
        ),
    color varchar(50) not null
);

create index idx_cars_owner_id on cars(owner_id);

-- =========================
-- 3. Услуги
-- =========================
create table services (
    service_id bigint generated always as identity primary key,
    name varchar(150) not null unique,
    description text,
    price_rub numeric(12,2) not null check (price_rub >= 0),
    regular_discount_percent numeric(5,2) not null default 0
        check (regular_discount_percent between 0 and 100),
    discounted_price_rub numeric(12,2)
        generated always as (
            round(price_rub * (100 - regular_discount_percent) / 100.0, 2)
        ) stored
);

-- =========================
-- 4. Сотрудники
-- =========================
create table employees (
    employee_id bigint generated always as identity primary key,
    personnel_number integer unique,
    specialty varchar(100) not null,
    phone varchar(20) not null unique,
    full_name varchar(200) not null unique
);

-- =========================
-- 5. Пользователи / аккаунты
-- Отдельная auth-сущность.
-- client -> owner
-- master -> employee
-- =========================
create table users (
    user_id bigint generated always as identity primary key,
    role varchar(20) not null
        check (role in ('client', 'master')),
    owner_id bigint unique references owners(owner_id)
        on update cascade
        on delete set null,
    employee_id bigint unique references employees(employee_id)
        on update cascade
        on delete set null,
    email varchar(255) not null unique,
    password_hash text not null,
    full_name varchar(200) not null,
    phone varchar(20) unique,
    preferred_entrypoint varchar(30) not null default 'profile'
        check (preferred_entrypoint in ('profile', 'cars', 'orders')),
    created_at timestamp without time zone not null default current_timestamp,
    updated_at timestamp without time zone not null default current_timestamp,
    last_login_at timestamp without time zone,
    constraint chk_users_role_link check (
        (role = 'client' and owner_id is not null and employee_id is null)
        or
        (role = 'master' and employee_id is not null and owner_id is null)
    )
);

create index idx_users_role on users(role);

-- =========================
-- 6. Зарплата сотрудников
-- 1:1 с сотрудником
-- =========================
create table employee_salaries (
    employee_id bigint primary key references employees(employee_id)
        on update cascade
        on delete cascade,
    participation_coeff numeric(4,2) not null
        check (participation_coeff > 0),
    full_shift_salary_rub numeric(12,2) not null
        check (full_shift_salary_rub >= 0),
    full_shift_salary_with_ktu_rub numeric(12,2)
        generated always as (
            round(full_shift_salary_rub * participation_coeff, 2)
        ) stored
);

-- =========================
-- 7. Детали для ремонта
-- =========================
create table repair_parts (
    repair_part_id bigint generated always as identity primary key,
    name varchar(150) not null unique,
    quantity integer not null check (quantity >= 0),
    delivery_date date
);

-- =========================
-- 8. Детали для ТО
-- =========================
create table maintenance_parts (
    maintenance_part_id bigint generated always as identity primary key,
    name varchar(150) not null unique,
    quantity integer not null check (quantity >= 0),
    delivery_date date
);

-- =========================
-- 9. Заказы автомастерской
-- =========================
create table orders (
    order_number bigint generated always as identity primary key,
    car_id bigint not null references cars(car_id)
        on update cascade
        on delete restrict,
    service_id bigint not null references services(service_id)
        on update cascade
        on delete restrict,
    ready_date date not null,
    created_at timestamp without time zone not null default current_timestamp
);

create index idx_orders_car_id on orders(car_id);
create index idx_orders_service_id on orders(service_id);

-- =========================
-- 10. Назначенный мастер в заказе
-- Оставляем имя таблицы order_employees,
-- чтобы не ломать существующие запросы backend.
-- Но теперь это связь 1 заказ -> 1 мастер.
-- =========================
create table order_employees (
    order_number bigint primary key references orders(order_number)
        on update cascade
        on delete cascade,
    employee_id bigint not null references employees(employee_id)
        on update cascade
        on delete restrict
);

create index idx_order_employees_employee_id on order_employees(employee_id);

-- =========================
-- 11. Детали для ремонта в заказе
-- =========================
create table order_repair_parts (
    order_number bigint not null references orders(order_number)
        on update cascade
        on delete cascade,
    repair_part_id bigint not null references repair_parts(repair_part_id)
        on update cascade
        on delete restrict,
    quantity_used integer not null default 1 check (quantity_used > 0),
    primary key (order_number, repair_part_id)
);

create index idx_order_repair_parts_part_id
    on order_repair_parts(repair_part_id);

-- =========================
-- 12. Детали ТО в заказе
-- =========================
create table order_maintenance_parts (
    order_number bigint not null references orders(order_number)
        on update cascade
        on delete cascade,
    maintenance_part_id bigint not null
        references maintenance_parts(maintenance_part_id)
        on update cascade
        on delete restrict,
    quantity_used integer not null default 1 check (quantity_used > 0),
    primary key (order_number, maintenance_part_id)
);

create index idx_order_maintenance_parts_part_id
    on order_maintenance_parts(maintenance_part_id);

-- =========================
-- 13. Представление, близкое к таблице
-- "Автомастерская" из РПЗ
-- =========================
create or replace view v_workshop_orders as
select
    o.order_number,
    c.plate_number as car_plate_number,
    ow.full_name as owner_full_name,
    ow.phone as owner_phone,
    s.name as service_name,
    o.ready_date,
    string_agg(distinct e.full_name, ', ') as employees
from orders o
join cars c on c.car_id = o.car_id
join owners ow on ow.owner_id = c.owner_id
join services s on s.service_id = o.service_id
left join order_employees oe on oe.order_number = o.order_number
left join employees e on e.employee_id = oe.employee_id
group by
    o.order_number,
    c.plate_number,
    ow.full_name,
    ow.phone,
    s.name,
    o.ready_date;

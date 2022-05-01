create table if not exists role (
    id bigserial unique not null constraint role_pkey primary key,
    nomination varchar(255) not null
);

comment on table role is 'Role in system';

insert into role values (1, 'initiator') on conflict do nothing;
insert into role values (2, 'secretary') on conflict do nothing;

create table if not exists "user" (
    id bigserial unique not null constraint user_pkey primary key,
    role_id bigint constraint role_fk references "role" on delete no action,
    email text not null,
    telegram_username varchar(255),
    chat_id bigint,
    created_at timestamp default now() not null
);

comment on table "user" is 'User data';

create table if not exists initiative (
    id bigserial unique not null constraint initiative_pkey primary key,
    user_id bigint not null constraint user_fk references "user" on delete no action,
    question text not null,
    yes int default 0 not null,
    no int default 0 not null,
    archive int default 0 not null,
    created_at timestamp default now() not null
);

comment on table initiative is 'Users initiative questions';
BEGIN;

    -- user points table
    create table if not exists user_points (
        user_id uuid references users(id) not null,
        points int not null,
        created_at timestamp default now(),
        updated_at timestamp,
        primary key (user_id)
    );

    -- user points history table
    create table if not exists user_points_history (
        user_points_history_id bigserial primary key,
        user_id uuid references users(id) not null,
        points int not null,
        created_at timestamp default now()
    );

    CREATE OR REPLACE FUNCTION public.on_user_points_change()
      RETURNS trigger
      LANGUAGE 'plpgsql'
    AS $BODY$
      BEGIN
        IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE'
        THEN
          EXECUTE
            'INSERT INTO user_points_history (user_id, points) VALUES ($1.user_id, $1.points)'
            USING NEW;
          RETURN NEW;
        END IF;
      END;
    $BODY$;

    CREATE TRIGGER user_points_history_trigger
      BEFORE INSERT OR UPDATE
      ON public.user_points
      FOR EACH ROW
      EXECUTE PROCEDURE public.on_user_points_change();

COMMIT;

CREATE TABLE "public"."users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "username" character varying NOT NULL,
  "password" character varying NOT NULL,
  "email" character varying NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_username_key" UNIQUE ("username"),
  CONSTRAINT "users_email_key" UNIQUE ("email")
);

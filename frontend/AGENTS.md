# Frontend Contributor Guide

## Structure

The Next.js 16+ application is under `frontend/`. Put routes and layouts in `app/`, shared React UI in `components/`, browser and API helpers in `lib/`, shared TypeScript types in `types/`, and static assets in `public/`. Place tests beside the code they cover, such as `components/StatusCard.test.tsx`.

## Commands

Run these commands from `frontend/`:

- `npm install` installs dependencies.
- `npm run dev` starts the local development server.
- `npm run build` creates a production build.
- `npm run lint` runs frontend linting.

## Style & Naming

Use two spaces in TypeScript and semicolons only where required. Name React components and types with PascalCase; name variables, functions, and helpers with camelCase. Keep reusable UI in `components/` and avoid placing unrelated data access in route components.

## Testing

Use the selected React testing framework and `*.test.ts` or `*.test.tsx` filenames. Add focused tests for new behavior. Before review, run the relevant tests plus `npm run lint` and `npm run build`.


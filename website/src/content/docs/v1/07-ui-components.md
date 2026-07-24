# 40+ Ejectable UI Components

Zyra includes 40+ accessible, customizable React UI components that require zero external npm runtime dependencies. All components support dark mode out of the box and seamlessly integrate with Zyra validation schemas.

## Ejectable Architecture

Unlike monolithic UI libraries that lock code inside `node_modules`, Zyra UI components are **ejectable**. Use `zyra add ui <component>` to copy clean component source code directly into your project's `components/ui/` folder for total customization.

```bash
zyra add ui button modal datatable form toast
```

---

## Component Showcase & Categories

### 1. Form & Input Components
- `<Button>` — Variant buttons with loading spinners & icons
- `<Input>` — Text, email, password fields with validation states
- `<Select>` — Accessible select dropdowns
- `<Checkbox>` & `<RadioGroup>` — Form state selectors
- `<DatePicker>` — Accessible date pickers
- `<ZyraForm>` — Automatic client + server validation bound to Go struct schemas

```tsx
import { ZyraForm, Input, Button } from '@/components/ui';
import { CreateUserInputSchema } from '@/.generated/schemas';

export function RegistrationForm() {
  return (
    <ZyraForm schema={CreateUserInputSchema} onSubmit={handleRegister}>
      <Input name="name" label="Full Name" />
      <Input name="email" label="Email Address" type="email" />
      <Input name="password" label="Password" type="password" />
      <Button type="submit">Create Account</Button>
    </ZyraForm>
  );
}
```

### 2. Feedback & Overlay Components
- `<Modal>` & `<Dialog>` — Accessible modal windows with focus trapping
- `<ToastProvider>` & `useToast()` — Animated notification toasts
- `<Alert>` — Success, warning, info, and error callouts
- `<Skeleton>` — Loading placeholder skeletons for `<ZyraBoundary>`

### 3. Navigation & Layout
- `<Sidebar>` & `<Navbar>` — Responsive navigation layouts
- `<Tabs>` — Keyboard accessible tab panels
- `<DropdownMenu>` — Context & user avatar menus
- `<Breadcrumb>` — Hierarchical navigation trails

### 4. Data Display
- `<DataTable>` — High-performance table with sorting, filtering, and pagination
- `<Badge>` — Status tags and pill indicators
- `<Avatar>` — User image and initials avatars
- `<Card>` — Surface containers with header, content, and footer slots

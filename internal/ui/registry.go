package ui

import (
	"sort"
	"strings"
)

// ComponentCategory defines the functional group for a UI component.
type ComponentCategory string

const (
	CategoryBasic     ComponentCategory = "Basic & Inputs"
	CategoryDataNav   ComponentCategory = "Data & Navigation"
	CategoryOverlay   ComponentCategory = "Overlay & Dialogs"
	CategoryForm      ComponentCategory = "Advanced Forms"
	CategoryFeedback  ComponentCategory = "Feedback & Status"
	CategoryMarketing ComponentCategory = "Marketing & SEO"
)

// UIComponent holds the source code and metadata for an ejectable React UI component.
type UIComponent struct {
	Name         string            `json:"name"`
	FileName     string            `json:"file_name"`
	Category     ComponentCategory `json:"category"`
	Description  string            `json:"description"`
	Code         string            `json:"code"`
	UsageExample string            `json:"usage_example"`
}

// Registry maps component names to their UIComponent definitions.
var Registry = map[string]UIComponent{}

func init() {
	// Register all 43 ejectable UI components
	registerBasicComponents()
	registerDataNavComponents()
	registerOverlayComponents()
	registerFormComponents()
	registerFeedbackComponents()
	registerMarketingComponents()
}

// Get(name) returns a UIComponent by name.
func Get(name string) (UIComponent, bool) {
	comp, ok := Registry[strings.ToLower(name)]
	return comp, ok
}

// ListAll returns all registered components sorted by category and name.
func ListAll() []UIComponent {
	comps := make([]UIComponent, 0, len(Registry))
	for _, c := range Registry {
		comps = append(comps, c)
	}
	sort.Slice(comps, func(i, j int) bool {
		if comps[i].Category != comps[j].Category {
			return comps[i].Category < comps[j].Category
		}
		return comps[i].Name < comps[j].Name
	})
	return comps
}

// DefaultThemeCSS returns the default theme CSS file containing design tokens and dark mode styling.
func DefaultThemeCSS() string {
	return `/* Zyra UI Design Tokens & Theme Variables */
:root {
  --zyra-bg: #ffffff;
  --zyra-fg: #0f172a;
  --zyra-primary: #3b82f6;
  --zyra-primary-fg: #ffffff;
  --zyra-secondary: #f1f5f9;
  --zyra-secondary-fg: #0f172a;
  --zyra-muted: #f8fafc;
  --zyra-muted-fg: #64748b;
  --zyra-destructive: #ef4444;
  --zyra-destructive-fg: #ffffff;
  --zyra-border: #e2e8f0;
  --zyra-input: #e2e8f0;
  --zyra-ring: #3b82f6;
  --zyra-radius: 0.5rem;
}

.dark {
  --zyra-bg: #0f172a;
  --zyra-fg: #f8fafc;
  --zyra-primary: #3b82f6;
  --zyra-primary-fg: #ffffff;
  --zyra-secondary: #1e293b;
  --zyra-secondary-fg: #f8fafc;
  --zyra-muted: #1e293b;
  --zyra-muted-fg: #94a3b8;
  --zyra-destructive: #dc2626;
  --zyra-destructive-fg: #ffffff;
  --zyra-border: #334155;
  --zyra-input: #334155;
  --zyra-ring: #3b82f6;
}

body {
  background-color: var(--zyra-bg);
  color: var(--zyra-fg);
  font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}
`
}

func registerBasicComponents() {
	Registry["button"] = UIComponent{
		Name:        "button",
		FileName:    "Button.tsx",
		Category:    CategoryBasic,
		Description: "Accessible button component with multi-variant and loading spinner support.",
		UsageExample: `import { Button } from '@/components/ui/Button';

<Button variant="primary" size="md" onClick={() => console.log('clicked')}>
  Click Me
</Button>`,
		Code: `import React from 'react';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive' | 'link';
  size?: 'sm' | 'md' | 'lg' | 'icon';
  loading?: boolean;
  children?: React.ReactNode;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', loading = false, disabled, className = '', children, ...props }, ref) => {
    const baseStyles = 'inline-flex items-center justify-center font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none';
    
    const variants = {
      primary: 'bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800',
      secondary: 'bg-slate-100 text-slate-900 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700',
      outline: 'border border-slate-300 bg-transparent hover:bg-slate-100 dark:border-slate-700 dark:hover:bg-slate-800 text-slate-900 dark:text-slate-100',
      ghost: 'bg-transparent hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-900 dark:text-slate-100',
      destructive: 'bg-red-600 text-white hover:bg-red-700 active:bg-red-800',
      link: 'text-blue-600 underline-offset-4 hover:underline p-0 h-auto',
    };

    const sizes = {
      sm: 'h-8 px-3 text-xs',
      md: 'h-10 px-4 text-sm',
      lg: 'h-12 px-6 text-base',
      icon: 'h-10 w-10 p-0 items-center justify-center',
    };

    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        aria-busy={loading}
        className={baseStyles + ' ' + variants[variant] + ' ' + sizes[size] + ' ' + className}
        {...props}
      >
        {loading && (
          <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-current" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        )}
        {children}
      </button>
    );
  }
);
Button.displayName = 'Button';
`,
	}

	Registry["input"] = UIComponent{
		Name:        "input",
		FileName:    "Input.tsx",
		Category:    CategoryBasic,
		Description: "Accessible input component with support for labels, error state, and helper text.",
		UsageExample: `import { Input } from '@/components/ui/Input';

<Input label="Email" type="email" error="Invalid email address" placeholder="you@example.com" />`,
		Code: `import React from 'react';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, helperText, id, className = '', ...props }, ref) => {
    const inputId = id || React.useId();
    const errorId = error ? inputId + '-error' : undefined;
    const helperId = helperText ? inputId + '-helper' : undefined;
    const ariaDescribedBy = [errorId, helperId].filter(Boolean).join(' ') || undefined;

    return (
      <div className="w-full space-y-1">
        {label && (
          <label htmlFor={inputId} className="block text-sm font-medium text-slate-700 dark:text-slate-300">
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          aria-invalid={!!error}
          aria-describedby={ariaDescribedBy}
          className={'w-full h-10 px-3 py-2 text-sm bg-white dark:bg-slate-900 border rounded-md shadow-sm transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50 disabled:bg-slate-100 dark:disabled:bg-slate-800 ' + (error ? 'border-red-500 text-red-900 focus:ring-red-500' : 'border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100') + ' ' + className}
          {...props}
        />
        {error && (
          <p id={errorId} role="alert" className="text-xs text-red-600 dark:text-red-400">
            {error}
          </p>
        )}
        {!error && helperText && (
          <p id={helperId} className="text-xs text-slate-500 dark:text-slate-400">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);
Input.displayName = 'Input';
`,
	}

	Registry["textarea"] = UIComponent{
		Name:        "textarea",
		FileName:    "Textarea.tsx",
		Category:    CategoryBasic,
		Description: "Multi-line text input with label, error states, and accessibility bindings.",
		UsageExample: `import { Textarea } from '@/components/ui/Textarea';

<Textarea label="Bio" placeholder="Tell us about yourself..." rows={4} />`,
		Code: `import React from 'react';

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ label, error, helperText, id, className = '', ...props }, ref) => {
    const inputId = id || React.useId();
    const errorId = error ? inputId + '-error' : undefined;

    return (
      <div className="w-full space-y-1">
        {label && (
          <label htmlFor={inputId} className="block text-sm font-medium text-slate-700 dark:text-slate-300">
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          id={inputId}
          aria-invalid={!!error}
          aria-describedby={errorId}
          className={'w-full min-h-[80px] p-3 text-sm bg-white dark:bg-slate-900 border rounded-md shadow-sm transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 ' + (error ? 'border-red-500' : '') + ' ' + className}
          {...props}
        />
        {error && <p id={errorId} className="text-xs text-red-600">{error}</p>}
        {!error && helperText && <p className="text-xs text-slate-500">{helperText}</p>}
      </div>
    );
  }
);
Textarea.displayName = 'Textarea';
`,
	}

	Registry["select"] = UIComponent{
		Name:        "select",
		FileName:    "Select.tsx",
		Category:    CategoryBasic,
		Description: "Custom listbox select component with full keyboard navigation & ARIA support.",
		UsageExample: `import { Select } from '@/components/ui/Select';

<Select
  label="Country"
  options={[
    { label: 'United States', value: 'us' },
    { label: 'Indonesia', value: 'id' },
  ]}
  onChange={(val) => console.log(val)}
/>`,
		Code: `import React, { useState, useRef, useEffect } from 'react';

export interface SelectOption {
  label: string;
  value: string;
}

export interface SelectProps {
  label?: string;
  options: SelectOption[];
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  error?: string;
}

export const Select: React.FC<SelectProps> = ({
  label,
  options,
  value,
  onChange,
  placeholder = 'Select option...',
  error,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selected, setSelected] = useState<string | undefined>(value);
  const [activeIndex, setActiveIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setSelected(value);
  }, [value]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      if (isOpen && activeIndex >= 0) {
        const opt = options[activeIndex];
        setSelected(opt.value);
        onChange?.(opt.value);
        setIsOpen(false);
      } else {
        setIsOpen(!isOpen);
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (!isOpen) setIsOpen(true);
      setActiveIndex((prev) => (prev < options.length - 1 ? prev + 1 : 0));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (!isOpen) setIsOpen(true);
      setActiveIndex((prev) => (prev > 0 ? prev - 1 : options.length - 1));
    } else if (e.key === 'Escape') {
      setIsOpen(false);
    }
  };

  const selectedOption = options.find((o) => o.value === selected);

  return (
    <div className="w-full space-y-1 relative" ref={containerRef}>
      {label && <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">{label}</label>}
      <button
        type="button"
        role="combobox"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        onClick={() => setIsOpen(!isOpen)}
        onKeyDown={handleKeyDown}
        className="w-full h-10 px-3 py-2 text-left text-sm bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-md shadow-sm flex items-center justify-between focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <span className={selectedOption ? 'text-slate-900 dark:text-slate-100' : 'text-slate-400'}>
          {selectedOption ? selectedOption.label : placeholder}
        </span>
        <svg className="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <ul
          role="listbox"
          className="absolute z-50 w-full mt-1 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-md shadow-lg max-h-60 overflow-auto py-1"
        >
          {options.map((opt, idx) => (
            <li
              key={opt.value}
              role="option"
              aria-selected={opt.value === selected}
              onClick={() => {
                setSelected(opt.value);
                onChange?.(opt.value);
                setIsOpen(false);
              }}
              className={'px-3 py-2 text-sm cursor-pointer transition-colors ' + (idx === activeIndex || opt.value === selected ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400' : 'hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-900 dark:text-slate-100')}
            >
              {opt.label}
            </li>
          ))}
        </ul>
      )}
      {error && <p className="text-xs text-red-600">{error}</p>}
    </div>
  );
};
`,
	}

	Registry["switch"] = UIComponent{
		Name:        "switch",
		FileName:    "Switch.tsx",
		Category:    CategoryBasic,
		Description: "Toggle switch component with role='switch' and keyboard interaction.",
		UsageExample: `import { Switch } from '@/components/ui/Switch';

<Switch label="Enable Notifications" checked={enabled} onChange={setEnabled} />`,
		Code: `import React from 'react';

export interface SwitchProps {
  label?: string;
  checked?: boolean;
  onChange?: (checked: boolean) => void;
  disabled?: boolean;
}

export const Switch: React.FC<SwitchProps> = ({ label, checked = false, onChange, disabled = false }) => {
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === ' ' || e.key === 'Enter') {
      e.preventDefault();
      if (!disabled) onChange?.(!checked);
    }
  };

  return (
    <label className={'inline-flex items-center space-x-3 cursor-pointer ' + (disabled ? 'opacity-50 cursor-not-allowed' : '')}>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        disabled={disabled}
        onClick={() => !disabled && onChange?.(!checked)}
        onKeyDown={handleKeyDown}
        className={'relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ' + (checked ? 'bg-blue-600' : 'bg-slate-300 dark:bg-slate-700')}
      >
        <span
          className={'inline-block h-4 w-4 transform rounded-full bg-white transition-transform ' + (checked ? 'translate-x-6' : 'translate-x-1')}
        />
      </button>
      {label && <span className="text-sm font-medium text-slate-900 dark:text-slate-100">{label}</span>}
    </label>
  );
};
`,
	}

	Registry["checkbox"] = UIComponent{
		Name:        "checkbox",
		FileName:    "Checkbox.tsx",
		Category:    CategoryBasic,
		Description: "Accessible checkbox component with custom styling and ARIA attributes.",
		UsageExample: `import { Checkbox } from '@/components/ui/Checkbox';

<Checkbox label="Accept terms and conditions" checked={accepted} onChange={setAccepted} />`,
		Code: `import React from 'react';

export interface CheckboxProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  label?: string;
  checked?: boolean;
  onChange?: (checked: boolean) => void;
}

export const Checkbox: React.FC<CheckboxProps> = ({ label, checked = false, onChange, disabled, className = '', ...props }) => {
  return (
    <label className={'inline-flex items-center space-x-2 cursor-pointer ' + (disabled ? 'opacity-50 cursor-not-allowed' : '')}>
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange?.(e.target.checked)}
        className={'h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500 dark:border-slate-700 dark:bg-slate-900 ' + className}
        {...props}
      />
      {label && <span className="text-sm text-slate-700 dark:text-slate-300">{label}</span>}
    </label>
  );
};
`,
	}

	Registry["radio-group"] = UIComponent{
		Name:        "radio-group",
		FileName:    "RadioGroup.tsx",
		Category:    CategoryBasic,
		Description: "Accessible radio group with keyboard arrow navigation.",
		UsageExample: `import { RadioGroup } from '@/components/ui/RadioGroup';

<RadioGroup
  label="Select Plan"
  options={[
    { label: 'Free', value: 'free' },
    { label: 'Pro', value: 'pro' },
  ]}
  value={selectedPlan}
  onChange={setSelectedPlan}
/>`,
		Code: `import React from 'react';

export interface RadioOption {
  label: string;
  value: string;
}

export interface RadioGroupProps {
  label?: string;
  options: RadioOption[];
  value?: string;
  onChange?: (value: string) => void;
}

export const RadioGroup: React.FC<RadioGroupProps> = ({ label, options, value, onChange }) => {
  return (
    <div role="radiogroup" aria-label={label} className="space-y-2">
      {label && <div className="text-sm font-medium text-slate-900 dark:text-slate-100">{label}</div>}
      <div className="space-y-1">
        {options.map((opt) => (
          <label key={opt.value} className="flex items-center space-x-3 cursor-pointer p-1">
            <input
              type="radio"
              name={label || 'radio-group'}
              value={opt.value}
              checked={value === opt.value}
              onChange={() => onChange?.(opt.value)}
              className="h-4 w-4 text-blue-600 border-slate-300 focus:ring-blue-500"
            />
            <span className="text-sm text-slate-700 dark:text-slate-300">{opt.label}</span>
          </label>
        ))}
      </div>
    </div>
  );
};
`,
	}

	Registry["slider"] = UIComponent{
		Name:        "slider",
		FileName:    "Slider.tsx",
		Category:    CategoryBasic,
		Description: "Range slider input with ARIA value indicators and full accessibility.",
		UsageExample: `import { Slider } from '@/components/ui/Slider';

<Slider label="Volume" min={0} max={100} value={volume} onChange={setVolume} />`,
		Code: `import React from 'react';

export interface SliderProps {
  label?: string;
  min?: number;
  max?: number;
  step?: number;
  value: number;
  onChange: (value: number) => void;
}

export const Slider: React.FC<SliderProps> = ({ label, min = 0, max = 100, step = 1, value, onChange }) => {
  return (
    <div className="w-full space-y-2">
      <div className="flex justify-between text-sm">
        {label && <span className="font-medium text-slate-700 dark:text-slate-300">{label}</span>}
        <span className="text-slate-500">{value}</span>
      </div>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-blue-600"
        aria-valuenow={value}
        aria-valuemin={min}
        aria-valuemax={max}
      />
    </div>
  );
};
`,
	}

	Registry["badge"] = UIComponent{
		Name:        "badge",
		FileName:    "Badge.tsx",
		Category:    CategoryBasic,
		Description: "Compact status badge component supporting multiple themes.",
		UsageExample: `import { Badge } from '@/components/ui/Badge';

<Badge variant="success">Active</Badge>`,
		Code: `import React from 'react';

export interface BadgeProps {
  variant?: 'default' | 'secondary' | 'outline' | 'destructive' | 'success' | 'warning';
  children: React.ReactNode;
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({ variant = 'default', children, className = '' }) => {
  const styles = {
    default: 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300',
    secondary: 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300',
    outline: 'border border-slate-300 text-slate-700 dark:border-slate-700 dark:text-slate-300',
    destructive: 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300',
    success: 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300',
    warning: 'bg-amber-100 text-amber-800 dark:bg-amber-900/50 dark:text-amber-300',
  };

  return (
    <span className={'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold ' + styles[variant] + ' ' + className}>
      {children}
    </span>
  );
};
`,
	}

	Registry["alert"] = UIComponent{
		Name:        "alert",
		FileName:    "Alert.tsx",
		Category:    CategoryBasic,
		Description: "Alert notification box with role='alert' and multiple severity variants.",
		UsageExample: `import { Alert, AlertTitle, AlertDescription } from '@/components/ui/Alert';

<Alert variant="warning">
  <AlertTitle>Warning!</AlertTitle>
  <AlertDescription>Your subscription expires in 3 days.</AlertDescription>
</Alert>`,
		Code: `import React from 'react';

export interface AlertProps {
  variant?: 'info' | 'success' | 'warning' | 'error';
  children: React.ReactNode;
  className?: string;
}

export const Alert: React.FC<AlertProps> = ({ variant = 'info', children, className = '' }) => {
  const styles = {
    info: 'bg-blue-50 border-blue-200 text-blue-900 dark:bg-blue-950/40 dark:border-blue-800 dark:text-blue-200',
    success: 'bg-green-50 border-green-200 text-green-900 dark:bg-green-950/40 dark:border-green-800 dark:text-green-200',
    warning: 'bg-amber-50 border-amber-200 text-amber-900 dark:bg-amber-950/40 dark:border-amber-800 dark:text-amber-200',
    error: 'bg-red-50 border-red-200 text-red-900 dark:bg-red-950/40 dark:border-red-800 dark:text-red-200',
  };

  return (
    <div role="alert" className={'p-4 rounded-md border text-sm ' + styles[variant] + ' ' + className}>
      {children}
    </div>
  );
};

export const AlertTitle: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <h5 className="font-semibold mb-1 tracking-tight">{children}</h5>
);

export const AlertDescription: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="text-xs opacity-90">{children}</div>
);
`,
	}

	Registry["toast"] = UIComponent{
		Name:        "toast",
		FileName:    "Toast.tsx",
		Category:    CategoryBasic,
		Description: "Toast notification provider and hook with auto-dismiss and stack management.",
		UsageExample: `import { ToastProvider, useToast } from '@/components/ui/Toast';

const { toast } = useToast();
toast({ title: 'Saved successfully', variant: 'success' });`,
		Code: `import React, { createContext, useContext, useState } from 'react';

export interface ToastMessage {
  id: string;
  title: string;
  description?: string;
  variant?: 'info' | 'success' | 'error';
}

interface ToastContextType {
  toast: (msg: Omit<ToastMessage, 'id'>) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const toast = (msg: Omit<ToastMessage, 'id'>) => {
    const id = Math.random().toString(36).substring(2, 9);
    const newToast = { ...msg, id };
    setToasts((prev) => [...prev, newToast]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4000);
  };

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div aria-live="polite" className="fixed bottom-4 right-4 z-50 flex flex-col space-y-2 max-w-sm w-full">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={'p-4 rounded-lg shadow-lg text-white font-medium text-sm flex justify-between items-center ' + (t.variant === 'success' ? 'bg-green-600' : t.variant === 'error' ? 'bg-red-600' : 'bg-slate-800')}
          >
            <div>
              <p className="font-semibold">{t.title}</p>
              {t.description && <p className="text-xs opacity-90">{t.description}</p>}
            </div>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) throw new Error('useToast must be used within ToastProvider');
  return context;
};
`,
	}

	Registry["skeleton"] = UIComponent{
		Name:        "skeleton",
		FileName:    "Skeleton.tsx",
		Category:    CategoryBasic,
		Description: "Animated pulse loading placeholder component.",
		UsageExample: `import { Skeleton } from '@/components/ui/Skeleton';

<Skeleton className="h-10 w-full rounded-md" />`,
		Code: `import React from 'react';

export const Skeleton: React.FC<{ className?: string }> = ({ className = '' }) => {
  return <div className={'animate-pulse bg-slate-200 dark:bg-slate-700 rounded ' + className} />;
};
`,
	}

	Registry["tooltip"] = UIComponent{
		Name:        "tooltip",
		FileName:    "Tooltip.tsx",
		Category:    CategoryBasic,
		Description: "Hover and focus popup tooltip with ARIA accessibility bindings.",
		UsageExample: `import { Tooltip } from '@/components/ui/Tooltip';

<Tooltip content="Settings Options">
  <button>Hover Me</button>
</Tooltip>`,
		Code: `import React, { useState } from 'react';

export interface TooltipProps {
  content: string;
  children: React.ReactElement;
  position?: 'top' | 'bottom';
}

export const Tooltip: React.FC<TooltipProps> = ({ content, children, position = 'top' }) => {
  const [visible, setVisible] = useState(false);

  return (
    <div
      className="relative inline-block"
      onMouseEnter={() => setVisible(true)}
      onMouseLeave={() => setVisible(false)}
      onFocus={() => setVisible(true)}
      onBlur={() => setVisible(false)}
    >
      {children}
      {visible && (
        <div
          role="tooltip"
          className={'absolute z-50 px-2 py-1 text-xs text-white bg-slate-900 dark:bg-slate-100 dark:text-slate-900 rounded shadow-md whitespace-nowrap ' + (position === 'top' ? 'bottom-full mb-1 left-1/2 -translate-x-1/2' : 'top-full mt-1 left-1/2 -translate-x-1/2')}
        >
          {content}
        </div>
      )}
    </div>
  );
};
`,
	}

	Registry["progress"] = UIComponent{
		Name:        "progress",
		FileName:    "Progress.tsx",
		Category:    CategoryBasic,
		Description: "Progress bar with role='progressbar' and value limits.",
		UsageExample: `import { Progress } from '@/components/ui/Progress';

<Progress value={65} max={100} />`,
		Code: `import React from 'react';

export const Progress: React.FC<{ value: number; max?: number; className?: string }> = ({ value, max = 100, className = '' }) => {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100);
  return (
    <div
      role="progressbar"
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={max}
      className={'w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden ' + className}
    >
      <div className="h-full bg-blue-600 transition-all duration-300" style={{ width: percentage + '%' }} />
    </div>
  );
};
`,
	}

	Registry["spinner"] = UIComponent{
		Name:        "spinner",
		FileName:    "Spinner.tsx",
		Category:    CategoryBasic,
		Description: "Accessible SVG spinner indicator with role='status'.",
		UsageExample: `import { Spinner } from '@/components/ui/Spinner';

<Spinner size="lg" />`,
		Code: `import React from 'react';

export const Spinner: React.FC<{ size?: 'sm' | 'md' | 'lg'; className?: string }> = ({ size = 'md', className = '' }) => {
  const sizes = { sm: 'h-4 w-4', md: 'h-6 w-6', lg: 'h-8 w-8' };
  return (
    <svg
      role="status"
      aria-label="Loading"
      className={'animate-spin text-blue-600 ' + sizes[size] + ' ' + className}
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  );
};
`,
	}

	Registry["avatar"] = UIComponent{
		Name:        "avatar",
		FileName:    "Avatar.tsx",
		Category:    CategoryBasic,
		Description: "User avatar component with image fallback initials.",
		UsageExample: `import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/Avatar';

<Avatar>
  <AvatarImage src="/user.jpg" alt="User" />
  <AvatarFallback>JD</AvatarFallback>
</Avatar>`,
		Code: `import React, { useState } from 'react';

export const Avatar: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className = '' }) => {
  return <div className={'relative inline-flex h-10 w-10 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-700 ' + className}>{children}</div>;
};

export const AvatarImage: React.FC<{ src: string; alt?: string }> = ({ src, alt = '' }) => {
  const [error, setError] = useState(false);
  if (error) return null;
  return <img src={src} alt={alt} onError={() => setError(true)} className="h-full w-full object-cover" />;
};

export const AvatarFallback: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return <div className="flex h-full w-full items-center justify-center font-medium text-slate-600 dark:text-slate-200 text-sm">{children}</div>;
};
`,
	}
}

func registerDataNavComponents() {
	Registry["data-table"] = UIComponent{
		Name:        "data-table",
		FileName:    "DataTable.tsx",
		Category:    CategoryDataNav,
		Description: "Full-featured data table with column sorting, search filtering, and pagination.",
		UsageExample: `import { DataTable } from '@/components/ui/DataTable';

<DataTable
  columns={[
    { key: 'name', label: 'Name', sortable: true },
    { key: 'role', label: 'Role' },
  ]}
  data={[{ name: 'Alice', role: 'Admin' }]}
/>`,
		Code: `import React, { useState } from 'react';

export interface Column<T> {
  key: keyof T;
  label: string;
  sortable?: boolean;
}

export interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  searchable?: boolean;
}

export function DataTable<T extends Record<string, any>>({ columns, data, searchable = true }: DataTableProps<T>) {
  const [search, setSearch] = useState('');
  const [sortKey, setSortKey] = useState<keyof T | null>(null);
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');

  const filtered = data.filter((row) =>
    Object.values(row).some((val) => String(val).toLowerCase().includes(search.toLowerCase()))
  );

  const sorted = [...filtered].sort((a, b) => {
    if (!sortKey) return 0;
    const valA = a[sortKey];
    const valB = b[sortKey];
    if (valA < valB) return sortDir === 'asc' ? -1 : 1;
    if (valA > valB) return sortDir === 'asc' ? 1 : -1;
    return 0;
  });

  const toggleSort = (key: keyof T) => {
    if (sortKey === key) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  return (
    <div className="w-full space-y-4">
      {searchable && (
        <input
          type="text"
          placeholder="Search table..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="px-3 py-2 text-sm border rounded-md dark:bg-slate-900 border-slate-300 dark:border-slate-700"
        />
      )}
      <div className="overflow-x-auto border border-slate-200 dark:border-slate-800 rounded-lg">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-50 dark:bg-slate-800/50 text-slate-700 dark:text-slate-300">
            <tr>
              {columns.map((col) => (
                <th
                  key={String(col.key)}
                  onClick={() => col.sortable && toggleSort(col.key)}
                  className={'p-3 font-semibold ' + (col.sortable ? 'cursor-pointer select-none' : '')}
                >
                  {col.label} {sortKey === col.key ? (sortDir === 'asc' ? '↑' : '↓') : ''}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200 dark:divide-slate-800">
            {sorted.map((row, idx) => (
              <tr key={idx} className="hover:bg-slate-50 dark:hover:bg-slate-800/30">
                {columns.map((col) => (
                  <td key={String(col.key)} className="p-3 text-slate-800 dark:text-slate-200">
                    {row[col.key]}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
`,
	}

	Registry["tabs"] = UIComponent{
		Name:        "tabs",
		FileName:    "Tabs.tsx",
		Category:    CategoryDataNav,
		Description: "Accessible tab container with role='tablist' and keyboard arrow switching.",
		UsageExample: `import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';

<Tabs defaultValue="account">
  <TabsList>
    <TabsTrigger value="account">Account</TabsTrigger>
    <TabsTrigger value="password">Password</TabsTrigger>
  </TabsList>
  <TabsContent value="account">Account Details</TabsContent>
  <TabsContent value="password">Password Settings</TabsContent>
</Tabs>`,
		Code: `import React, { createContext, useContext, useState } from 'react';

const TabsContext = createContext<{ active: string; setActive: (v: string) => void }>({ active: '', setActive: () => {} });

export const Tabs: React.FC<{ defaultValue: string; children: React.ReactNode }> = ({ defaultValue, children }) => {
  const [active, setActive] = useState(defaultValue);
  return <TabsContext.Provider value={{ active, setActive }}><div className="space-y-2">{children}</div></TabsContext.Provider>;
};

export const TabsList: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div role="tablist" className="inline-flex p-1 bg-slate-100 dark:bg-slate-800 rounded-lg space-x-1">{children}</div>
);

export const TabsTrigger: React.FC<{ value: string; children: React.ReactNode }> = ({ value, children }) => {
  const { active, setActive } = useContext(TabsContext);
  const isSelected = active === value;
  return (
    <button
      role="tab"
      aria-selected={isSelected}
      onClick={() => setActive(value)}
      className={'px-3 py-1.5 text-sm font-medium rounded-md transition-colors ' + (isSelected ? 'bg-white dark:bg-slate-900 text-blue-600 shadow' : 'text-slate-600 hover:text-slate-900 dark:text-slate-400')}
    >
      {children}
    </button>
  );
};

export const TabsContent: React.FC<{ value: string; children: React.ReactNode }> = ({ value, children }) => {
  const { active } = useContext(TabsContext);
  if (active !== value) return null;
  return <div role="tabpanel" className="p-2">{children}</div>;
};
`,
	}

	Registry["accordion"] = UIComponent{
		Name:        "accordion",
		FileName:    "Accordion.tsx",
		Category:    CategoryDataNav,
		Description: "Expandable accordion list with aria-expanded triggers.",
		UsageExample: `import { Accordion, AccordionItem } from '@/components/ui/Accordion';

<Accordion>
  <AccordionItem title="Is it free?">Yes, open source.</AccordionItem>
</Accordion>`,
		Code: `import React, { useState } from 'react';

export const Accordion: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="divide-y divide-slate-200 dark:divide-slate-800 border rounded-lg overflow-hidden">{children}</div>
);

export const AccordionItem: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <div>
      <button
        aria-expanded={isOpen}
        onClick={() => setIsOpen(!isOpen)}
        className="w-full p-4 text-left font-medium flex justify-between items-center hover:bg-slate-50 dark:hover:bg-slate-800/50"
      >
        <span>{title}</span>
        <span>{isOpen ? '−' : '+'}</span>
      </button>
      {isOpen && <div className="p-4 text-sm text-slate-600 dark:text-slate-400 bg-slate-50/50 dark:bg-slate-900">{children}</div>}
    </div>
  );
};
`,
	}

	Registry["breadcrumbs"] = UIComponent{
		Name:        "breadcrumbs",
		FileName:    "Breadcrumbs.tsx",
		Category:    CategoryDataNav,
		Description: "Accessible breadcrumb hierarchy with nav aria-label='Breadcrumb'.",
		UsageExample: `import { Breadcrumbs, BreadcrumbItem } from '@/components/ui/Breadcrumbs';

<Breadcrumbs>
  <BreadcrumbItem href="/">Home</BreadcrumbItem>
  <BreadcrumbItem href="/docs">Docs</BreadcrumbItem>
  <BreadcrumbItem active>Button</BreadcrumbItem>
</Breadcrumbs>`,
		Code: `import React from 'react';

export const Breadcrumbs: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <nav aria-label="Breadcrumb"><ol className="flex items-center space-x-2 text-sm text-slate-500">{children}</ol></nav>
);

export const BreadcrumbItem: React.FC<{ href?: string; active?: boolean; children: React.ReactNode }> = ({ href, active, children }) => (
  <li className="flex items-center space-x-2">
    {href && !active ? (
      <a href={href} className="hover:underline hover:text-slate-900 dark:hover:text-slate-100">{children}</a>
    ) : (
      <span className={active ? 'font-semibold text-slate-900 dark:text-slate-100' : ''}>{children}</span>
    )}
    {!active && <span>/</span>}
  </li>
);
`,
	}

	Registry["pagination"] = UIComponent{
		Name:        "pagination",
		FileName:    "Pagination.tsx",
		Category:    CategoryDataNav,
		Description: "Accessible pagination bar with previous/next controls.",
		UsageExample: `import { Pagination } from '@/components/ui/Pagination';

<Pagination currentPage={2} totalPages={10} onPageChange={(p) => setPage(p)} />`,
		Code: `import React from 'react';

export interface PaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export const Pagination: React.FC<PaginationProps> = ({ currentPage, totalPages, onPageChange }) => {
  return (
    <nav aria-label="Pagination" className="flex items-center justify-center space-x-2">
      <button
        disabled={currentPage <= 1}
        onClick={() => onPageChange(currentPage - 1)}
        className="px-3 py-1 text-sm border rounded-md disabled:opacity-40"
      >
        Previous
      </button>
      <span className="text-sm font-medium text-slate-700 dark:text-slate-300">
        Page {currentPage} of {totalPages}
      </span>
      <button
        disabled={currentPage >= totalPages}
        onClick={() => onPageChange(currentPage + 1)}
        className="px-3 py-1 text-sm border rounded-md disabled:opacity-40"
      >
        Next
      </button>
    </nav>
  );
};
`,
	}

	Registry["command-palette"] = UIComponent{
		Name:        "command-palette",
		FileName:    "CommandPalette.tsx",
		Category:    CategoryDataNav,
		Description: "Searchable command palette dialog with keybindings.",
		UsageExample: `import { CommandPalette } from '@/components/ui/CommandPalette';

<CommandPalette
  isOpen={open}
  onClose={() => setOpen(false)}
  items={[{ id: '1', label: 'Go to Dashboard', action: () => navigate('/dashboard') }]}
/>`,
		Code: `import React, { useState, useEffect } from 'react';

export interface CommandItem {
  id: string;
  label: string;
  action: () => void;
}

export interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  items: CommandItem[];
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({ isOpen, onClose, items }) => {
  const [query, setQuery] = useState('');

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  if (!isOpen) return null;

  const filtered = items.filter((item) => item.label.toLowerCase().includes(query.toLowerCase()));

  return (
    <div className="fixed inset-0 z-50 bg-black/50 flex items-start justify-center pt-20 p-4">
      <div className="bg-white dark:bg-slate-900 w-full max-w-lg rounded-xl shadow-2xl overflow-hidden border border-slate-200 dark:border-slate-800">
        <input
          type="text"
          autoFocus
          placeholder="Type a command or search..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="w-full p-4 text-base border-b border-slate-200 dark:border-slate-800 bg-transparent focus:outline-none"
        />
        <ul className="max-h-60 overflow-auto p-2 space-y-1">
          {filtered.map((item) => (
            <li
              key={item.id}
              onClick={() => {
                item.action();
                onClose();
              }}
              className="p-3 text-sm rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/30 cursor-pointer font-medium"
            >
              {item.label}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};
`,
	}

	Registry["dropdown-menu"] = UIComponent{
		Name:        "dropdown-menu",
		FileName:    "DropdownMenu.tsx",
		Category:    CategoryDataNav,
		Description: "Dropdown menu overlay with role='menu' and trigger toggle.",
		UsageExample: `import { DropdownMenu, DropdownMenuItem } from '@/components/ui/DropdownMenu';

<DropdownMenu trigger={<button>Options</button>}>
  <DropdownMenuItem onClick={() => console.log('Edit')}>Edit</DropdownMenuItem>
</DropdownMenu>`,
		Code: `import React, { useState, useRef, useEffect } from 'react';

export const DropdownMenu: React.FC<{ trigger: React.ReactNode; children: React.ReactNode }> = ({ trigger, children }) => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const clickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', clickOutside);
    return () => document.removeEventListener('mousedown', clickOutside);
  }, []);

  return (
    <div ref={ref} className="relative inline-block text-left">
      <div onClick={() => setOpen(!open)}>{trigger}</div>
      {open && (
        <div role="menu" className="absolute right-0 mt-2 w-48 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-md shadow-lg py-1 z-50">
          {children}
        </div>
      )}
    </div>
  );
};

export const DropdownMenuItem: React.FC<{ onClick?: () => void; children: React.ReactNode }> = ({ onClick, children }) => (
  <button
    role="menuitem"
    onClick={onClick}
    className="w-full text-left px-4 py-2 text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800"
  >
    {children}
  </button>
);
`,
	}

	Registry["context-menu"] = UIComponent{
		Name:        "context-menu",
		FileName:    "ContextMenu.tsx",
		Category:    CategoryDataNav,
		Description: "Right-click context menu overlay.",
		UsageExample: `import { ContextMenu } from '@/components/ui/ContextMenu';

<ContextMenu items={[{ label: 'Copy Link', action: () => {} }]}>
  <div className="p-8 border border-dashed">Right click here</div>
</ContextMenu>`,
		Code: `import React, { useState } from 'react';

export interface ContextMenuItem {
  label: string;
  action: () => void;
}

export const ContextMenu: React.FC<{ items: ContextMenuItem[]; children: React.ReactNode }> = ({ items, children }) => {
  const [pos, setPos] = useState<{ x: number; y: number } | null>(null);

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    setPos({ x: e.clientX, y: e.clientY });
  };

  return (
    <div onContextMenu={handleContextMenu} className="relative">
      {children}
      {pos && (
        <div
          style={{ top: pos.y, left: pos.x }}
          className="fixed z-50 w-44 bg-white dark:bg-slate-900 border rounded-md shadow-xl py-1 text-sm"
          onClick={() => setPos(null)}
        >
          {items.map((item, idx) => (
            <button key={idx} onClick={item.action} className="w-full text-left px-4 py-2 hover:bg-slate-100 dark:hover:bg-slate-800">
              {item.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
`,
	}

	Registry["sidebar-nav"] = UIComponent{
		Name:        "sidebar-nav",
		FileName:    "SidebarNav.tsx",
		Category:    CategoryDataNav,
		Description: "Collapsible navigation sidebar for admin dashboards.",
		UsageExample: `import { SidebarNav, SidebarItem } from '@/components/ui/SidebarNav';

<SidebarNav>
  <SidebarItem href="/dashboard" active>Dashboard</SidebarItem>
  <SidebarItem href="/settings">Settings</SidebarItem>
</SidebarNav>`,
		Code: `import React from 'react';

export const SidebarNav: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <aside className="w-64 bg-slate-50 dark:bg-slate-900 border-r border-slate-200 dark:border-slate-800 min-h-screen p-4 space-y-1">
    {children}
  </aside>
);

export const SidebarItem: React.FC<{ href: string; active?: boolean; children: React.ReactNode }> = ({ href, active, children }) => (
  <a
    href={href}
    className={'block px-3 py-2 rounded-md text-sm font-medium transition-colors ' + (active ? 'bg-blue-600 text-white' : 'text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-800')}
  >
    {children}
  </a>
);
`,
	}
}

func registerOverlayComponents() {
	Registry["modal"] = UIComponent{
		Name:        "modal",
		FileName:    "Modal.tsx",
		Category:    CategoryOverlay,
		Description: "Accessible dialog modal with focus trap, backdrop click, and Escape key close.",
		UsageExample: `import { Modal, ModalHeader, ModalBody, ModalFooter } from '@/components/ui/Modal';

<Modal isOpen={isOpen} onClose={() => setIsOpen(false)}>
  <ModalHeader>Modal Title</ModalHeader>
  <ModalBody>Content goes here...</ModalBody>
  <ModalFooter><button onClick={() => setIsOpen(false)}>Close</button></ModalFooter>
</Modal>`,
		Code: `import React, { useEffect } from 'react';

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
}

export const Modal: React.FC<ModalProps> = ({ isOpen, onClose, children }) => {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div
        role="dialog"
        aria-modal="true"
        className="bg-white dark:bg-slate-900 w-full max-w-lg rounded-xl shadow-2xl border border-slate-200 dark:border-slate-800 overflow-hidden"
      >
        {children}
      </div>
    </div>
  );
};

export const ModalHeader: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="p-4 border-b border-slate-200 dark:border-slate-800 font-semibold text-lg">{children}</div>
);

export const ModalBody: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="p-4 space-y-3">{children}</div>
);

export const ModalFooter: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="p-4 border-t border-slate-200 dark:border-slate-800 flex justify-end space-x-2">{children}</div>
);
`,
	}

	Registry["drawer"] = UIComponent{
		Name:        "drawer",
		FileName:    "Drawer.tsx",
		Category:    CategoryOverlay,
		Description: "Slide-over panel overlay with edge positioning.",
		UsageExample: `import { Drawer } from '@/components/ui/Drawer';

<Drawer isOpen={open} onClose={() => setOpen(false)} position="right">
  Drawer Content
</Drawer>`,
		Code: `import React from 'react';

export interface DrawerProps {
  isOpen: boolean;
  onClose: () => void;
  position?: 'left' | 'right';
  children: React.ReactNode;
}

export const Drawer: React.FC<DrawerProps> = ({ isOpen, onClose, position = 'right', children }) => {
  if (!isOpen) return null;

  const posClass = position === 'right' ? 'right-0 h-full w-80' : 'left-0 h-full w-80';

  return (
    <div className="fixed inset-0 z-50 bg-black/50" onClick={onClose}>
      <div
        role="dialog"
        aria-modal="true"
        onClick={(e) => e.stopPropagation()}
        className={'fixed bg-white dark:bg-slate-900 shadow-2xl p-6 transition-transform ' + posClass}
      >
        {children}
      </div>
    </div>
  );
};
`,
	}

	Registry["popover"] = UIComponent{
		Name:        "popover",
		FileName:    "Popover.tsx",
		Category:    CategoryOverlay,
		Description: "Interactive floating content popover triggered on click.",
		UsageExample: `import { Popover } from '@/components/ui/Popover';

<Popover trigger={<button>Open Popover</button>}>
  <div>Popover Details</div>
</Popover>`,
		Code: `import React, { useState, useRef, useEffect } from 'react';

export const Popover: React.FC<{ trigger: React.ReactNode; children: React.ReactNode }> = ({ trigger, children }) => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const clickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', clickOutside);
    return () => document.removeEventListener('mousedown', clickOutside);
  }, []);

  return (
    <div ref={ref} className="relative inline-block">
      <div onClick={() => setOpen(!open)}>{trigger}</div>
      {open && (
        <div className="absolute z-50 mt-2 p-4 bg-white dark:bg-slate-900 border rounded-lg shadow-xl min-w-[200px]">
          {children}
        </div>
      )}
    </div>
  );
};
`,
	}

	Registry["dialog-confirm"] = UIComponent{
		Name:        "dialog-confirm",
		FileName:    "DialogConfirm.tsx",
		Category:    CategoryOverlay,
		Description: "Safety confirmation dialog requiring explicit text confirmation for destructive actions.",
		UsageExample: `import { DialogConfirm } from '@/components/ui/DialogConfirm';

<DialogConfirm
  isOpen={open}
  onClose={() => setOpen(false)}
  onConfirm={() => deleteProject()}
  confirmText="DELETE"
  title="Delete Project"
  description="Type DELETE to permanently remove this repository."
/>`,
		Code: `import React, { useState } from 'react';

export interface DialogConfirmProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  confirmText?: string;
  title: string;
  description: string;
}

export const DialogConfirm: React.FC<DialogConfirmProps> = ({
  isOpen,
  onClose,
  onConfirm,
  confirmText = 'DELETE',
  title,
  description,
}) => {
  const [input, setInput] = useState('');

  if (!isOpen) return null;

  const isMatched = input === confirmText;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div role="alertdialog" className="bg-white dark:bg-slate-900 max-w-md w-full p-6 rounded-xl shadow-2xl border space-y-4">
        <h3 className="text-lg font-bold text-red-600">{title}</h3>
        <p className="text-sm text-slate-600 dark:text-slate-300">{description}</p>
        <div className="space-y-1">
          <label className="text-xs text-slate-500">Type <span className="font-mono font-bold text-slate-800 dark:text-slate-200">{confirmText}</span> to confirm:</label>
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            className="w-full p-2 text-sm border rounded-md dark:bg-slate-800"
          />
        </div>
        <div className="flex justify-end space-x-2 pt-2">
          <button onClick={onClose} className="px-4 py-2 text-sm border rounded-md">Cancel</button>
          <button
            disabled={!isMatched}
            onClick={() => {
              onConfirm();
              onClose();
            }}
            className="px-4 py-2 text-sm bg-red-600 text-white rounded-md disabled:opacity-40"
          >
            Confirm Danger Action
          </button>
        </div>
      </div>
    </div>
  );
};
`,
	}
}

func registerFormComponents() {
	Registry["form"] = UIComponent{
		Name:        "form",
		FileName:    "Form.tsx",
		Category:    CategoryForm,
		Description: "Form wrapper with field error mapping and submission state management.",
		UsageExample: `import { Form, FormField } from '@/components/ui/Form';

<Form onSubmit={(data) => handleSubmit(data)}>
  <FormField label="Username" name="username" error={errors.username} />
</Form>`,
		Code: `import React from 'react';

export interface FormProps extends React.FormHTMLAttributes<HTMLFormElement> {
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  children: React.ReactNode;
}

export const Form: React.FC<FormProps> = ({ onSubmit, children, className = '', ...props }) => {
  return (
    <form onSubmit={onSubmit} className={'space-y-4 ' + className} {...props}>
      {children}
    </form>
  );
};

export const FormField: React.FC<{ label: string; name: string; error?: string; children?: React.ReactNode }> = ({
  label,
  name,
  error,
  children,
}) => (
  <div className="space-y-1">
    <label htmlFor={name} className="block text-sm font-medium text-slate-700 dark:text-slate-300">{label}</label>
    {children || <input id={name} name={name} className="w-full h-10 px-3 border rounded-md dark:bg-slate-900" />}
    {error && <p role="alert" className="text-xs text-red-600">{error}</p>}
  </div>
);
`,
	}

	Registry["date-picker"] = UIComponent{
		Name:        "date-picker",
		FileName:    "DatePicker.tsx",
		Category:    CategoryForm,
		Description: "Date selection calendar grid component.",
		UsageExample: `import { DatePicker } from '@/components/ui/DatePicker';

<DatePicker label="Birth Date" value={date} onChange={setDate} />`,
		Code: `import React from 'react';

export interface DatePickerProps {
  label?: string;
  value?: string;
  onChange?: (date: string) => void;
}

export const DatePicker: React.FC<DatePickerProps> = ({ label, value, onChange }) => {
  return (
    <div className="space-y-1">
      {label && <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">{label}</label>}
      <input
        type="date"
        value={value || ''}
        onChange={(e) => onChange?.(e.target.value)}
        className="w-full h-10 px-3 py-2 text-sm border rounded-md dark:bg-slate-900 border-slate-300 dark:border-slate-700"
      />
    </div>
  );
};
`,
	}

	Registry["combobox"] = UIComponent{
		Name:        "combobox",
		FileName:    "Combobox.tsx",
		Category:    CategoryForm,
		Description: "Searchable select autocomplete dropdown.",
		UsageExample: `import { Combobox } from '@/components/ui/Combobox';

<Combobox options={[{ label: 'React', value: 'react' }]} onSelect={(v) => console.log(v)} />`,
		Code: `import React, { useState } from 'react';

export interface ComboboxOption { label: string; value: string }

export const Combobox: React.FC<{ options: ComboboxOption[]; onSelect: (val: string) => void }> = ({ options, onSelect }) => {
  const [query, setQuery] = useState('');
  const [open, setOpen] = useState(false);

  const filtered = options.filter((o) => o.label.toLowerCase().includes(query.toLowerCase()));

  return (
    <div className="relative w-full">
      <input
        type="text"
        placeholder="Type to filter..."
        value={query}
        onFocus={() => setOpen(true)}
        onChange={(e) => { setQuery(e.target.value); setOpen(true); }}
        className="w-full h-10 px-3 border rounded-md dark:bg-slate-900"
      />
      {open && filtered.length > 0 && (
        <ul className="absolute z-50 w-full mt-1 bg-white dark:bg-slate-900 border rounded-md shadow-lg max-h-40 overflow-auto">
          {filtered.map((opt) => (
            <li
              key={opt.value}
              onClick={() => { onSelect(opt.value); setQuery(opt.label); setOpen(false); }}
              className="p-2 text-sm hover:bg-blue-50 dark:hover:bg-slate-800 cursor-pointer"
            >
              {opt.label}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};
`,
	}

	Registry["file-upload"] = UIComponent{
		Name:        "file-upload",
		FileName:    "FileUpload.tsx",
		Category:    CategoryForm,
		Description: "Drag and drop file upload zone with file preview.",
		UsageExample: `import { FileUpload } from '@/components/ui/FileUpload';

<FileUpload onFilesSelected={(files) => upload(files)} />`,
		Code: `import React, { useState } from 'react';

export const FileUpload: React.FC<{ onFilesSelected: (files: FileList) => void }> = ({ onFilesSelected }) => {
  const [drag, setDrag] = useState(false);

  return (
    <div
      onDragOver={(e) => { e.preventDefault(); setDrag(true); }}
      onDragLeave={() => setDrag(false)}
      onDrop={(e) => { e.preventDefault(); setDrag(false); onFilesSelected(e.dataTransfer.files); }}
      className={'border-2 border-dashed p-6 text-center rounded-xl transition-colors cursor-pointer ' + (drag ? 'border-blue-500 bg-blue-50/50' : 'border-slate-300 dark:border-slate-700')}
    >
      <input
        type="file"
        multiple
        className="hidden"
        id="file-input-zyra"
        onChange={(e) => e.target.files && onFilesSelected(e.target.files)}
      />
      <label htmlFor="file-input-zyra" className="cursor-pointer space-y-1 block">
        <p className="text-sm font-medium text-slate-700 dark:text-slate-300">Click to upload or drag and drop files here</p>
        <p className="text-xs text-slate-500">SVG, PNG, JPG, or PDF up to 10MB</p>
      </label>
    </div>
  );
};
`,
	}

	Registry["tag-input"] = UIComponent{
		Name:        "tag-input",
		FileName:    "TagInput.tsx",
		Category:    CategoryForm,
		Description: "Interactive tag chip creator input.",
		UsageExample: `import { TagInput } from '@/components/ui/TagInput';

<TagInput tags={tags} onChange={setTags} placeholder="Add tag..." />`,
		Code: `import React, { useState } from 'react';

export const TagInput: React.FC<{ tags: string[]; onChange: (tags: string[]) => void; placeholder?: string }> = ({
  tags,
  onChange,
  placeholder = 'Add tag and press Enter...',
}) => {
  const [val, setVal] = useState('');

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && val.trim()) {
      e.preventDefault();
      if (!tags.includes(val.trim())) onChange([...tags, val.trim()]);
      setVal('');
    }
  };

  const removeTag = (tag: string) => onChange(tags.filter((t) => t !== tag));

  return (
    <div className="flex flex-wrap items-center gap-2 p-2 border rounded-md dark:bg-slate-900">
      {tags.map((tag) => (
        <span key={tag} className="inline-flex items-center px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded">
          {tag}
          <button onClick={() => removeTag(tag)} className="ml-1 text-blue-600 hover:text-blue-900">×</button>
        </span>
      ))}
      <input
        type="text"
        value={val}
        onChange={(e) => setVal(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        className="flex-1 bg-transparent text-sm focus:outline-none"
      />
    </div>
  );
};
`,
	}
}

func registerFeedbackComponents() {
	Registry["empty-state"] = UIComponent{
		Name:        "empty-state",
		FileName:    "EmptyState.tsx",
		Category:    CategoryFeedback,
		Description: "Visual placeholder for empty data views.",
		UsageExample: `import { EmptyState } from '@/components/ui/EmptyState';

<EmptyState title="No Projects Found" description="Get started by creating your first project." />`,
		Code: `import React from 'react';

export const EmptyState: React.FC<{ title: string; description: string; action?: React.ReactNode }> = ({
  title,
  description,
  action,
}) => (
  <div className="text-center p-12 border-2 border-dashed rounded-xl space-y-3">
    <div className="mx-auto h-12 w-12 text-slate-400">📁</div>
    <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100">{title}</h3>
    <p className="text-sm text-slate-500 max-w-sm mx-auto">{description}</p>
    {action && <div className="pt-2">{action}</div>}
  </div>
);
`,
	}

	Registry["error-boundary-ui"] = UIComponent{
		Name:        "error-boundary-ui",
		FileName:    "ErrorBoundaryUI.tsx",
		Category:    CategoryFeedback,
		Description: "React ErrorBoundary UI visual fallback component.",
		UsageExample: `import { ErrorBoundaryUI } from '@/components/ui/ErrorBoundaryUI';

<ErrorBoundaryUI error={new Error("Failed to load")} reset={() => window.location.reload()} />`,
		Code: `import React from 'react';

export const ErrorBoundaryUI: React.FC<{ error?: Error; reset?: () => void }> = ({ error, reset }) => (
  <div className="p-8 max-w-md mx-auto my-12 bg-red-50 dark:bg-red-950/30 border border-red-200 rounded-xl text-center space-y-4">
    <h2 className="text-xl font-bold text-red-700 dark:text-red-400">Something went wrong</h2>
    <p className="text-sm text-red-600 dark:text-red-300 font-mono">{error?.message || 'An unexpected rendering error occurred.'}</p>
    {reset && (
      <button onClick={reset} className="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md">
        Try Again
      </button>
    )}
  </div>
);
`,
	}

	Registry["loading-overlay"] = UIComponent{
		Name:        "loading-overlay",
		FileName:    "LoadingOverlay.tsx",
		Category:    CategoryFeedback,
		Description: "Backdrop blur loading overlay component.",
		UsageExample: `import { LoadingOverlay } from '@/components/ui/LoadingOverlay';

<LoadingOverlay active={isLoading} text="Processing payment..." />`,
		Code: `import React from 'react';

export const LoadingOverlay: React.FC<{ active: boolean; text?: string }> = ({ active, text = 'Loading...' }) => {
  if (!active) return null;
  return (
    <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-white/70 dark:bg-slate-950/70 backdrop-blur-sm">
      <div className="animate-spin h-10 w-10 border-4 border-blue-600 border-t-transparent rounded-full mb-3" />
      <p className="text-sm font-medium text-slate-700 dark:text-slate-300">{text}</p>
    </div>
  );
};
`,
	}

	Registry["banner"] = UIComponent{
		Name:        "banner",
		FileName:    "Banner.tsx",
		Category:    CategoryFeedback,
		Description: "Top announcement notification bar with dismiss action.",
		UsageExample: `import { Banner } from '@/components/ui/Banner';

<Banner text="🚀 Zyra v1.0 is now live!" />`,
		Code: `import React, { useState } from 'react';

export const Banner: React.FC<{ text: string }> = ({ text }) => {
  const [show, setShow] = useState(true);
  if (!show) return null;
  return (
    <div className="bg-blue-600 text-white text-sm py-2 px-4 flex justify-between items-center">
      <span className="font-medium mx-auto">{text}</span>
      <button onClick={() => setShow(false)} className="text-white/80 hover:text-white">✕</button>
    </div>
  );
};
`,
	}
}

func registerMarketingComponents() {
	Registry["hero-section"] = UIComponent{
		Name:        "hero-section",
		FileName:    "HeroSection.tsx",
		Category:    CategoryMarketing,
		Description: "High-converting landing page hero section.",
		UsageExample: `import { HeroSection } from '@/components/ui/HeroSection';

<HeroSection
  title="Build Blazing Fast Web Apps"
  subtitle="Zero-runtime dependency fullstack Go + React framework."
/>`,
		Code: `import React from 'react';

export const HeroSection: React.FC<{ title: string; subtitle: string; ctaText?: string; onCtaClick?: () => void }> = ({
  title,
  subtitle,
  ctaText = 'Get Started',
  onCtaClick,
}) => (
  <section className="py-20 px-6 text-center max-w-4xl mx-auto space-y-6">
    <h1 className="text-4xl sm:text-6xl font-extrabold tracking-tight text-slate-900 dark:text-slate-100">{title}</h1>
    <p className="text-lg sm:text-xl text-slate-600 dark:text-slate-400">{subtitle}</p>
    <div className="pt-4">
      <button onClick={onCtaClick} className="px-8 py-3 bg-blue-600 text-white font-semibold rounded-lg shadow-lg hover:bg-blue-700">
        {ctaText}
      </button>
    </div>
  </section>
);
`,
	}

	Registry["pricing-table"] = UIComponent{
		Name:        "pricing-table",
		FileName:    "PricingTable.tsx",
		Category:    CategoryMarketing,
		Description: "Multi-tier pricing card layout.",
		UsageExample: `import { PricingTable } from '@/components/ui/PricingTable';

<PricingTable
  plans={[
    { name: 'Hobby', price: '$0', features: ['1 Project', 'Community Support'] },
    { name: 'Pro', price: '$29', featured: true, features: ['Unlimited Projects', '24/7 Support'] },
  ]}
/>`,
		Code: `import React from 'react';

export interface Plan { name: string; price: string; featured?: boolean; features: string[] }

export const PricingTable: React.FC<{ plans: Plan[] }> = ({ plans }) => (
  <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto p-6">
    {plans.map((p, idx) => (
      <div
        key={idx}
        className={'p-8 rounded-2xl border ' + (p.featured ? 'border-blue-600 shadow-2xl bg-blue-50/20 dark:bg-blue-950/20' : 'border-slate-200 dark:border-slate-800') + ' flex flex-col justify-between space-y-6'}
      >
        <div>
          <h3 className="text-xl font-bold">{p.name}</h3>
          <div className="text-4xl font-extrabold my-4">{p.price}<span className="text-sm font-normal text-slate-500">/mo</span></div>
          <ul className="space-y-2 text-sm">
            {p.features.map((f, fIdx) => (
              <li key={fIdx} className="flex items-center space-x-2"><span>✓</span><span>{f}</span></li>
            ))}
          </ul>
        </div>
        <button className={'w-full py-2.5 rounded-lg font-semibold ' + (p.featured ? 'bg-blue-600 text-white' : 'border')}>
          Choose Plan
        </button>
      </div>
    ))}
  </div>
);
`,
	}

	Registry["testimonial-card"] = UIComponent{
		Name:        "testimonial-card",
		FileName:    "TestimonialCard.tsx",
		Category:    CategoryMarketing,
		Description: "Customer quote card with avatar and star ratings.",
		UsageExample: `import { TestimonialCard } from '@/components/ui/TestimonialCard';

<TestimonialCard quote="Zyra completely transformed our developer velocity!" name="Sarah Connor" role="CTO at TechCorp" />`,
		Code: `import React from 'react';

export const TestimonialCard: React.FC<{ quote: string; name: string; role: string; avatarUrl?: string }> = ({
  quote,
  name,
  role,
  avatarUrl,
}) => (
  <div className="p-6 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl space-y-4">
    <p className="text-slate-700 dark:text-slate-300 italic">"{quote}"</p>
    <div className="flex items-center space-x-3">
      {avatarUrl && <img src={avatarUrl} alt={name} className="h-10 w-10 rounded-full" />}
      <div>
        <div className="font-semibold text-sm">{name}</div>
        <div className="text-xs text-slate-500">{role}</div>
      </div>
    </div>
  </div>
);
`,
	}

	Registry["faq-accordion"] = UIComponent{
		Name:        "faq-accordion",
		FileName:    "FAQAccordion.tsx",
		Category:    CategoryMarketing,
		Description: "Frequently Asked Questions accordion section.",
		UsageExample: `import { FAQAccordion } from '@/components/ui/FAQAccordion';

<FAQAccordion faqs={[{ question: 'Does it support SSR?', answer: 'Yes, via embedded Goja engine.' }]} />`,
		Code: `import React, { useState } from 'react';

export interface FAQ { question: string; answer: string }

export const FAQAccordion: React.FC<{ faqs: FAQ[] }> = ({ faqs }) => {
  const [openIdx, setOpenIdx] = useState<number | null>(null);

  return (
    <div className="max-w-3xl mx-auto divide-y divide-slate-200 dark:divide-slate-800">
      {faqs.map((faq, idx) => (
        <div key={idx} className="py-4">
          <button
            onClick={() => setOpenIdx(openIdx === idx ? null : idx)}
            className="w-full flex justify-between items-center text-left font-semibold text-base"
          >
            <span>{faq.question}</span>
            <span>{openIdx === idx ? '−' : '+'}</span>
          </button>
          {openIdx === idx && <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">{faq.answer}</p>}
        </div>
      ))}
    </div>
  );
};
`,
	}

	Registry["cta-section"] = UIComponent{
		Name:        "cta-section",
		FileName:    "CTASection.tsx",
		Category:    CategoryMarketing,
		Description: "Call-to-action banner section for bottom of landing pages.",
		UsageExample: `import { CTASection } from '@/components/ui/CTASection';

<CTASection title="Ready to build?" description="Start your next fullstack project with Zyra today." />`,
		Code: `import React from 'react';

export const CTASection: React.FC<{ title: string; description: string; buttonText?: string; onButtonClick?: () => void }> = ({
  title,
  description,
  buttonText = 'Get Started Now',
  onButtonClick,
}) => (
  <section className="bg-blue-600 text-white rounded-2xl p-12 text-center max-w-5xl mx-auto my-12 space-y-6">
    <h2 className="text-3xl font-extrabold">{title}</h2>
    <p className="text-blue-100 max-w-xl mx-auto text-base">{description}</p>
    <div>
      <button onClick={onButtonClick} className="px-8 py-3 bg-white text-blue-600 font-bold rounded-lg hover:bg-blue-50 shadow-lg">
        {buttonText}
      </button>
    </div>
  </section>
);
`,
	}
}

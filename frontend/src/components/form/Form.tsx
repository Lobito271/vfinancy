import type { ReactNode } from 'react';
import { FormProvider, useForm, type UseFormProps, type UseFormReturn, type FieldValues, type SubmitHandler } from 'react-hook-form';

export interface FormProps<T extends FieldValues> extends Omit<UseFormProps<T>, 'children'> {
  onSubmit: SubmitHandler<T>;
  children: ReactNode | ((form: UseFormReturn<T>) => ReactNode);
}

export function Form<T extends FieldValues>({ onSubmit, children, ...props }: FormProps<T>) {
  const form = useForm<T>(props);
  return (
    <FormProvider {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {typeof children === 'function' ? children(form) : children}
      </form>
    </FormProvider>
  );
}

import type { ReactNode } from 'react';
import {
  FormProvider,
  useForm,
  type UseFormProps,
  type UseFormReturn,
  type FieldValues,
  type FieldErrors,
  type SubmitHandler,
  type Resolver,
} from 'react-hook-form';
import { z } from 'zod';

interface FormProps<T extends FieldValues>
  extends Omit<UseFormProps<T, any, T>, 'children' | 'resolver'> {
  schema?: z.ZodType<T>;
  onSubmit: SubmitHandler<T>;
  children: ReactNode | ((form: UseFormReturn<T, any, T>) => ReactNode);
}

function makeResolver<T extends FieldValues>(schema: z.ZodType<T>): Resolver<T, any, T> {
  return (values) => {
    const result = schema.safeParse(values);
    if (result.success) {
      return { values: result.data as T, errors: {} };
    }
    const errors: Record<string, { type: string; message: string }> = {};
    for (const issue of result.error.issues) {
      const path = issue.path.length ? issue.path.join('.') : 'root';
      if (!errors[path]) {
        errors[path] = { type: issue.code, message: issue.message };
      }
    }
    return { values: {}, errors: errors as unknown as FieldErrors<T> };
  };
}

export function Form<T extends FieldValues>({ schema, onSubmit, children, ...props }: FormProps<T>) {
  const form = useForm<T, any, T>({ ...props, resolver: schema ? makeResolver(schema) : undefined });
  return (
    <FormProvider {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="form" noValidate>
        {typeof children === 'function' ? children(form) : children}
      </form>
    </FormProvider>
  );
}

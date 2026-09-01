import * as React from 'react';
import { Tabs as TabsPrimitive } from '@base-ui/react/tabs';
import { cx } from '@/utils/cx';

export const Tabs = TabsPrimitive.Root;

export const TabsList = React.forwardRef<
  React.ComponentRef<typeof TabsPrimitive.List>,
  Omit<React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>, 'className'> & { className?: string }
>(({ className, ...props }, ref) => (
  <TabsPrimitive.List ref={ref} className={cx('tabs-list', className)} {...props} />
));
TabsList.displayName = 'TabsList';

export const TabsTrigger = React.forwardRef<
  React.ComponentRef<typeof TabsPrimitive.Tab>,
  Omit<React.ComponentPropsWithoutRef<typeof TabsPrimitive.Tab>, 'className'> & { className?: string }
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Tab ref={ref} className={cx('tabs-trigger', className)} {...props} />
));
TabsTrigger.displayName = 'TabsTrigger';

export const TabsContent = React.forwardRef<
  React.ComponentRef<typeof TabsPrimitive.Panel>,
  Omit<React.ComponentPropsWithoutRef<typeof TabsPrimitive.Panel>, 'className'> & { className?: string }
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Panel ref={ref} className={cx('tabs-content', className)} {...props} />
));
TabsContent.displayName = 'TabsContent';

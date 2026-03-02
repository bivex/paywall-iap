"use client";

import { TableRow } from "@/components/ui/table";
import { TransactionDetailSheet } from "./transaction-detail-sheet";
import type { TransactionRow } from "@/actions/transactions";

interface Props {
  tx: TransactionRow;
  children: React.ReactNode;
}

export function TxRow({ tx, children }: Props) {
  return (
    <TransactionDetailSheet
      transactionId={tx.id}
      trigger={
        <TableRow className="hover:bg-muted/30 cursor-pointer">
          {children}
        </TableRow>
      }
    />
  );
}

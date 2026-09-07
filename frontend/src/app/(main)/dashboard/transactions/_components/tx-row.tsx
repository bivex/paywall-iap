"use client";

import type { TransactionRow } from "@/actions/transactions";
import { TableRow } from "@/components/ui/table";

import { TransactionDetailSheet } from "./transaction-detail-sheet";

interface Props {
  tx: TransactionRow;
  children: React.ReactNode;
}

export function TxRow({ tx, children }: Props) {
  return (
    <TransactionDetailSheet
      transactionId={tx.id}
      trigger={<TableRow className="hover:bg-muted/30 cursor-pointer">{children}</TableRow>}
    />
  );
}

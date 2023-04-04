package org.example;

import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

class SearchFiles {

    public static void main(String[] args) {
        if (args.length < 2) {
            System.out.println("Запускай так: java SearchFiles <path> <word> [threads]");
            return;
        }

        String path = args[0];
        SearchRunner.findWord = args[1].toLowerCase();
        int threads = 1;
        if (args.length == 3) {
            threads = toInt(args[2]);
            if (threads <= 0) {
                System.out.println("Невалидное кол-во тредов");
                return;
            }
        }


        long startTime = System.nanoTime();

        SearchRunner.queue.add(path);
        SearchRunner.executor = Executors.newFixedThreadPool(threads);
        for (int i = 0; i < threads; i++) {
            SearchRunner.executor.submit(new SearchRunner());
        }

        try {
            SearchRunner.executor.awaitTermination(Long.MAX_VALUE, TimeUnit.NANOSECONDS);
        } catch (InterruptedException e) {
            throw new RuntimeException(e);
        }

        long duration = System.nanoTime() - startTime;
        double seconds = duration / 1e9;

        System.out.println("Нашёл " + SearchRunner.ans.size() + " файл(ов) за " + seconds + " секунд:");
        for (String file : SearchRunner.ans) {
            System.out.println(file);
        }
    }

    public static int toInt(String s) {
        try {
            return Integer.parseInt(s);
        } catch (NumberFormatException e) {
            return -1;
        }
    }

}
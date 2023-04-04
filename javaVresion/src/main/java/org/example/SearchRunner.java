package org.example;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.io.File;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.LinkedBlockingQueue;


public class SearchRunner implements Runnable {
    public static ExecutorService executor;
    final Logger log = LogManager.getLogger(SearchRunner.class);
    public final static BlockingQueue<String> queue = new LinkedBlockingQueue<>();
    public static List<String> ans = new ArrayList<>();
    public static String findWord;
    public static Boolean isFinishMain = false;
    public static final Boolean block = false;

    @Override
    public void run() {
        log.info("Start thread!");
        search();
        log.info("End thread!");

        // если очередь пустая, то поток заканчивается,
        // но может оказатся, что очередь не успела заполнится.
        // Это регулируем через поток номер 1
        synchronized (block) {
            if (Thread.currentThread().getName().contains("thread-1")) {
                executor.shutdown();
                isFinishMain = true;
            } else {
                if (!isFinishMain) {
                    executor.submit(new SearchRunner());
                }
            }
        }
    }

    private void search() {
        while (true) {
            String source = queue.poll();
            if (source == null) {
                return;
            }

            for (File child : Objects.requireNonNull(new File(source).listFiles())) {
                if (child.isDirectory()) {
                    try {
                        queue.put(child.getAbsolutePath());
                    } catch (InterruptedException e) {
                        throw new RuntimeException(e);
                    }
                }

                if (child.getName().toLowerCase().contains(findWord)) {
                    ans.add(child.getAbsolutePath());
                }
            }
        }
    }

}

                         db
                        d88b       ,d      ,d
                       d8'`8b      88      88
                      d8'  `8b   MM88MMM MM88MMM ,adPPYba,
                     d8YaaaaY8b    88      88   a8"     "8a
                    d8""""""""8b   88      88   8b       d8
                   d8'        `8b  88,     88,  "8a,   ,a8"
                  d8'          `8b "Y888   "Y888 `"YbbdP"'

+------------------------------------------------------------------------------+

1.  About

    Atto is a lightweight, opinionated text editor written in Go. The current
    feature set is quite limited, but Atto is being actively developed. Many
    features and improvements should be added in the coming weeks.

2.  Features

    The following features are currently available in Atto:

      - Core editor functionality (text viewing & editing)
      - Multiple simultaneous buffers
      - Simple syntax highlighting (Go & C)
      - User configuration files (options limited)

    In addition to the features above, the following features are planned:

      - Copy/cut/paste functionality
      - Smarter syntax highlighting with user-definable language syntax files

    NOTICE: Atto is developed and tested on macOS, but should also work on Linux
    and BSD systems. Windows is not supported at this time.

3.  Installation

    Just build and place the binary somewhere in your PATH.

4.  Usage

    $ atto <file>

    For help using Atto, a list of shortcuts & usage information can be shown
    by pressing ^H inside Atto.

5.  Configuration

    When you start Atto for the first time, a configuration folder will be
    created for you at '~/.atto'. Inside you will find a config.yml file which
    you can edit to change the editor's exposed preferences.

6.  License

    Atto is licensed under the MIT License. See LICENSE.txt for more info.
